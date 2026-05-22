package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/internal/api/auth"
	"github.com/insmtx/Leros/backend/internal/api/contract"
	"github.com/insmtx/Leros/backend/internal/infra/db"
	eventbus "github.com/insmtx/Leros/backend/internal/infra/mq"
	"github.com/insmtx/Leros/backend/internal/worker/protocol"
	"github.com/insmtx/Leros/backend/pkg/dm"
	"github.com/insmtx/Leros/backend/types"
	"github.com/ygpkg/yg-go/encryptor/snowflake"
	"github.com/ygpkg/yg-go/logs"
)

var _ contract.WorkService = (*workService)(nil)

type workService struct {
	db       *gorm.DB
	eventbus eventbus.EventBus
	inferrer AssistantInferrer
}

func NewWorkService(database *gorm.DB, eventbus eventbus.EventBus, inferrer AssistantInferrer) contract.WorkService {
	return &workService{
		db:       database,
		eventbus: eventbus,
		inferrer: inferrer,
	}
}

func (s *workService) NewMessage(ctx context.Context, req *contract.NewMessageRequest) (*contract.NewMessageResponse, error) {
	if req.Content == "" {
		return nil, errors.New("content is required")
	}

	caller, _ := auth.FromContext(ctx)
	if caller == nil || caller.Uin == 0 || caller.OrgID == 0 {
		return nil, errors.New("user not authenticated or org not set")
	}

	c := &newMessageCtx{s: s, ctx: ctx, req: req, caller: caller}

	if err := c.resolveOrCreateProject(); err != nil {
		return nil, err
	}
	if err := c.ensureProjectSession(); err != nil {
		return nil, err
	}
	if err := c.resolveOrCreateTask(); err != nil {
		return nil, err
	}
	if err := c.createTaskSession(); err != nil {
		return nil, err
	}
	if err := c.createMessage(); err != nil {
		return nil, err
	}
	if err := c.publishMessageEvents(); err != nil {
		return nil, err
	}
	if err := c.publishWorkerTask(); err != nil {
		return nil, err
	}

	return &contract.NewMessageResponse{
		ProjectID:   c.project.PublicID,
		TaskID:      c.task.PublicID,
		SessionID:   c.taskSession.PublicID,
		MessageID:   fmt.Sprintf("%d", c.message.ID),
		AssistantID: c.taskSession.AllocatedAssistantID,
	}, nil
}

type newMessageCtx struct {
	s      *workService
	ctx    context.Context
	req    *contract.NewMessageRequest
	caller *types.Caller

	project     *types.Project
	task        *types.Task
	taskSession *types.Session
	message     *types.SessionMessage
}

func (c *newMessageCtx) resolveOrCreateProject() error {
	if c.req.ProjectID != "" {
		p, err := db.GetProjectByPublicID(c.ctx, c.s.db, c.caller.OrgID, c.req.ProjectID)
		if err != nil {
			return err
		}
		if p == nil {
			return errors.New("project not found")
		}
		c.project = p
		return nil
	}

	runes := []rune(c.req.Content)
	title := string(runes)
	if len(runes) > 50 {
		title = string(runes[:50])
	}

	projectID := fmt.Sprintf("prj_%s", snowflake.GenerateIDBase58())
	c.project = &types.Project{
		PublicID:    projectID,
		OrgID:       c.caller.OrgID,
		OwnerID:     c.caller.Uin,
		Name:        title,
		Description: "",
		Status:      string(types.ProjectStatusActive),
	}
	if err := db.CreateProject(c.ctx, c.s.db, c.project); err != nil {
		return fmt.Errorf("create project: %w", err)
	}

	if err := db.CreateProjectMember(c.ctx, c.s.db, &types.ProjectMember{
		ProjectID:  c.project.ID,
		MemberID:   c.caller.Uin,
		MemberType: types.MemberTypeUser,
		MemberRole: types.MemberRoleOwner,
	}); err != nil {
		logs.WarnContextf(c.ctx, "create project member failed: %v", err)
	}

	return nil
}

func (c *newMessageCtx) ensureProjectSession() error {
	projectSession, err := db.GetProjectSession(c.ctx, c.s.db, c.project.ID)
	if err != nil {
		return fmt.Errorf("get project session: %w", err)
	}
	if projectSession != nil {
		return nil
	}

	projectSessionID := fmt.Sprintf("sess_%s", snowflake.GenerateIDBase58())
	projectSession = &types.Session{
		PublicID:             projectSessionID,
		Type:                 types.SessionTypeProject,
		Uin:                  c.caller.Uin,
		OrgID:                c.caller.OrgID,
		AssistantID:          c.req.AssistantID,
		AllocatedAssistantID: c.req.AssistantID,
		ProjectID:            &c.project.ID,
		Status:               string(types.SessionStatusActive),
		Title:                "项目协作",
	}
	if err := db.CreateSession(c.ctx, c.s.db, projectSession); err != nil {
		return fmt.Errorf("create project session: %w", err)
	}
	return nil
}

func (c *newMessageCtx) resolveOrCreateTask() error {
	if c.req.TaskID != "" {
		t, err := db.GetTaskByPublicID(c.ctx, c.s.db, c.req.TaskID)
		if err != nil {
			return err
		}
		if t == nil {
			return errors.New("task not found")
		}
		c.task = t
		return nil
	}

	runes := []rune(c.req.Content)
	taskTitle := string(runes)
	if len(runes) > 50 {
		taskTitle = string(runes[:50])
	}

	taskID := fmt.Sprintf("task_%s", snowflake.GenerateIDBase58())
	c.task = &types.Task{
		PublicID:    taskID,
		OrgID:       c.caller.OrgID,
		OwnerID:     c.caller.Uin,
		ProjectID:   c.project.ID,
		TaskType:    types.TaskTypeGeneral,
		Title:       taskTitle,
		Description: c.req.Content,
		Status:      string(types.TaskStatusCreated),
	}
	if err := db.CreateTask(c.ctx, c.s.db, c.task); err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	return nil
}

func (c *newMessageCtx) createTaskSession() error {
	taskSessionID := fmt.Sprintf("sess_%s", snowflake.GenerateIDBase58())
	c.taskSession = &types.Session{
		PublicID:             taskSessionID,
		Type:                 types.SessionTypeTask,
		Uin:                  c.caller.Uin,
		OrgID:                c.caller.OrgID,
		AssistantID:          c.req.AssistantID,
		AllocatedAssistantID: c.req.AssistantID,
		ProjectID:            &c.project.ID,
		TaskID:               &c.task.ID,
		Status:               string(types.SessionStatusActive),
		Title:                c.task.Title,
	}
	if err := db.CreateSession(c.ctx, c.s.db, c.taskSession); err != nil {
		return fmt.Errorf("create task session: %w", err)
	}

	c.task.SessionID = &c.taskSession.ID
	if err := c.s.db.WithContext(c.ctx).Model(c.task).Update("session_id", c.taskSession.ID).Error; err != nil {
		logs.WarnContextf(c.ctx, "update task session_id failed: %v", err)
	}

	return nil
}

func (c *newMessageCtx) createMessage() error {
	sequence, err := db.GetNextSequence(c.ctx, c.s.db, c.taskSession.ID)
	if err != nil {
		return err
	}

	msgType := c.req.MessageType
	if msgType == "" {
		msgType = string(types.MessageTypeText)
	}

	c.message = &types.SessionMessage{
		SessionID:   c.taskSession.ID,
		Role:        string(types.MessageRoleUser),
		Content:     c.req.Content,
		MessageType: msgType,
		Status:      string(types.MessageStatusPending),
		Sequence:    sequence,
		Timestamp:   time.Now().UnixMilli(),
	}
	if err := db.CreateMessage(c.ctx, c.s.db, c.message); err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	return nil
}

func (c *newMessageCtx) publishMessageEvents() error {
	now := time.Now()
	if err := db.IncrementMessageCount(c.ctx, c.s.db, c.taskSession.ID); err != nil {
		return err
	}
	if err := db.UpdateLastMessageAt(c.ctx, c.s.db, c.taskSession.ID, now); err != nil {
		return err
	}

	if c.taskSession.OrgID > 0 {
		topic, err := dm.SessionMessageRequestSubject(c.taskSession.OrgID, c.taskSession.PublicID)
		if err != nil {
			logs.WarnContextf(c.ctx, "failed to build message request subject: %v", err)
		} else {
			if err := c.s.eventbus.Publish(c.ctx, topic, c.message); err != nil {
				logs.WarnContextf(c.ctx, "failed to publish message to eventbus: %v", err)
			}
		}
	}

	return nil
}

func (c *newMessageCtx) publishWorkerTask() error {
	return c.s.publishWorkerTask(c.ctx, c.taskSession, c.message)
}

func (s *workService) publishWorkerTask(ctx context.Context, session *types.Session, message *types.SessionMessage) error {
	caller, _ := auth.FromContext(ctx)
	orgID := session.OrgID
	if orgID == 0 && caller != nil {
		orgID = caller.OrgID
	}

	if session.AssistantID == 0 && session.AllocatedAssistantID == 0 && s.inferrer != nil {
		assignedAssistantID := s.inferrer.InferAssignedAssistantID(ctx, orgID, string(session.Type))
		if assignedAssistantID > 0 {
			session.AllocatedAssistantID = assignedAssistantID
			if err := db.UpdateAllocatedAssistantID(ctx, s.db, session.ID, assignedAssistantID); err != nil {
				return fmt.Errorf("failed to update allocated_assistant_id: %w", err)
			}
		}
	}

	if session.AllocatedAssistantID == 0 {
		logs.DebugContextf(ctx, "Skipping task publish: no worker allocated for session %s", session.PublicID)
		return nil
	}

	topic, err := dm.WorkerTaskSubject(orgID, session.AllocatedAssistantID)
	if err != nil {
		return fmt.Errorf("failed to construct worker task topic: %w", err)
	}

	messagePayload := protocol.WorkerTaskMessage{
		ID:        fmt.Sprintf("msg_%d_%d", session.ID, message.Sequence),
		Type:      protocol.MessageTypeWorkerTask,
		CreatedAt: time.Now().UTC(),
		Trace: protocol.TraceContext{
			TraceID:   session.PublicID,
			RequestID: fmt.Sprintf("req_%d", message.ID),
			TaskID:    fmt.Sprintf("task_%d", message.ID),
		},
		Route: protocol.RouteContext{
			OrgID:     orgID,
			SessionID: session.PublicID,
			WorkerID:  session.AllocatedAssistantID,
		},
		Body: protocol.WorkerTaskBody{
			TaskType: protocol.TaskTypeAgentRun,
			Actor: protocol.ActorContext{
				UserID:      fmt.Sprintf("%d", session.Uin),
				DisplayName: "",
				Channel:     "session",
			},
			Input: protocol.TaskInput{
				Type: protocol.InputTypeMessage,
			},
		},
		Metadata: map[string]any{
			"session_id":   session.PublicID,
			"message_type": message.MessageType,
			"sequence":     message.Sequence,
			"timestamp":    message.Timestamp,
		},
	}

	if err := s.eventbus.Publish(ctx, topic, messagePayload); err != nil {
		logs.ErrorContextf(ctx, "Failed to publish message to assistant %d: %v", session.AllocatedAssistantID, err)
		return fmt.Errorf("failed to publish message to assistant: %w", err)
	}
	logs.DebugContextf(ctx, "Published message to topic %s: session_id=%s sequence=%d", topic, session.PublicID, message.Sequence)
	return nil
}
