"use client";

import type { NavItem, Project, ViewMode } from "@leros/store";
import { useLayoutStore } from "@leros/store";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@leros/ui/components/ui/dropdown-menu";
import { ScrollArea } from "@leros/ui/components/ui/scroll-area";
import { cn } from "@leros/ui/lib/utils";
import {
	ChevronDown,
	CircleHelp,
	Hash,
	LayoutGrid,
	LogOut,
	Network,
	Settings,
	ClipboardList,
	Zap,
	Database,
	ChevronsLeft,
	MoreVertical,
	UserRound,
} from "lucide-react";

const avatarMap: Record<string, string> = {
	"Ada AI": "https://lh3.googleusercontent.com/aida-public/AB6AXuDFpBbS4l95muQqtwMYtUuf8WCwNc5sA8OO0-6u1LGuYyluoaArOURURsMTCrMq_NupAuGHz-JOO1FokisXhPwW2YHHw98AiRCPLBB7pnEkJtJ49IFY1oAvXh91Jm-_COCvYzzzLBiaLG-LYG1u2FkKZ0I32-W4xkWSIw9t0g-REw0_7AApPcTHTUs6YXhMUR8CRrgkQwLTEXmTGIXKdTeB49LdA0NLB84cpa3IeofhyuLdIwA_DqEbSLLGdzjPLvMzaF8LprQnlCI",
	"Hopper": "https://lh3.googleusercontent.com/aida-public/AB6AXuBeB5b4oXNn4L2BxiToWnXKcmpiqIOQXHgzr--j9T9_QOXVd9oHi1Fm6w-TFVrtUCrsljLwuZTLgUsQO_bm-5a-pTeEhYiqC-XWGCFm29XVQNzs1K_BZsauTofNldKOlXXqefrOEws7yf2OugGY02bc3tTG6Ar6LK_vtTM0LIGPIUtjF4hXiV6_JC78AZjUIIcQ9ZyIsXqZHT4w005HdcD-k2UMVDi9B4zKpMqsRbKjO_uJgC-cMhnEekpNM3Tao6dm5c2dEHGt1m4",
	"Mia": "https://lh3.googleusercontent.com/aida-public/AB6AXuBF0owbtXZ299YjKA9U1M8sCOv64scrlTj0dggJ4QzZ3LVWiwaw6F2wdlx-pfng186UXwb39pUr6UYaB3TR0VgvyCzHeq_ftW0GiYK6opisJR6rW9cI41epBVwQ01amJW2zeCfuSC4bO9eHQmG3birvJfEvqhddLBP9UAyGwjti4KWyfS5HGYrOGMI1T2aGvaWbAMOO-dYq22Ezmpl3PWzyb7yd1yYy2LEOqAOSuhmadQKH90cgkhBTISnC5mE8jOrwmrdZuF-Fvs4"
};

const iconMap: Record<string, React.ReactNode> = {
	IconWorkbench: <LayoutGrid className="size-5" />,
	IconTask: <ClipboardList className="size-5" />,
	IconSkill: <Zap className="size-5" />,
	IconKnowledge: <Database className="size-5" />,
	IconProject: <Hash className="size-4" />,
};

const navIdToView: Record<string, ViewMode> = {
	workbench: "workbench",
	tasks: "tasks",
	knowledge: "knowledge",
	skills: "skills",
	"ai-1": "digitalAssistant",
	"ai-2": "digitalAssistant",
	"ai-3": "digitalAssistant",
};

export function LeftRail({ logoSrc = "/logo.svg" }: { logoSrc?: string }) {
	const { navGroups, projects, currentView, activeProjectId, switchView, switchProject } =
		useLayoutStore((s) => s);

	const handleNavClick = (item: NavItem) => {
		const view = navIdToView[item.id] ?? "chat";
		switchView(view);
	};

	const isItemActive = (item: NavItem) => {
		const view = navIdToView[item.id] ?? "chat";
		return currentView === view;
	};

	return (
		<aside className="leros-sidebar">
			<div className="leros-brand">
				<div className="flex items-center gap-3">
					<div className="leros-logo-placeholder bg-[#0050cb] rounded-xl flex items-center justify-center text-white shadow-sm w-10 h-10" aria-hidden="true">
						<Network className="size-5" />
					</div>
					<div className="min-w-0">
						<div className="leros-brand-title font-bold text-slate-900 leading-tight text-[18px]">Leros AI</div>
						<div className="leros-brand-version text-[10px] text-slate-400 font-medium mt-[2px]">v0.1</div>
					</div>
				</div>
				<button type="button" className="text-slate-300 hover:text-slate-600 transition-colors" aria-label="收起侧边栏">
					<ChevronsLeft className="size-[18px]" />
				</button>
			</div>

			<ScrollArea className="min-h-0 flex-1 overflow-hidden">
				<nav className="leros-nav" aria-label="主导航">
					{navGroups.map((group) => {
						return (
							<div key={group.id} className="leros-nav-section">
								{group.label && <div className="leros-nav-section-label">{group.label}</div>}
								{group.id === "projects" ? (
									<ProjectList
										projects={projects}
										activeProjectId={activeProjectId}
										currentView={currentView}
										onProjectClick={switchProject}
									/>
								) : (
									<div className="space-y-1">
										{group.items.map((item: NavItem) => (
											<NavItemButton
												key={item.id}
												item={item}
												active={isItemActive(item)}
												onClick={() => handleNavClick(item)}
											/>
										))}
									</div>
								)}
							</div>
						);
					})}
				</nav>
			</ScrollArea>

			<div className="leros-sidebar-footer shrink-0">
				<DropdownMenu>
					<DropdownMenuTrigger
						render={
							<button type="button" className="leros-profile-trigger flex items-center gap-3 px-3 py-3 rounded-xl hover:bg-slate-200/50 transition-all">
								<span className="leros-avatar overflow-hidden size-10 rounded-full border border-slate-200 object-cover flex-shrink-0">
									<img
										src="https://lh3.googleusercontent.com/aida-public/AB6AXuBF0owbtXZ299YjKA9U1M8sCOv64scrlTj0dggJ4QzZ3LVWiwaw6F2wdlx-pfng186UXwb39pUr6UYaB3TR0VgvyCzHeq_ftW0GiYK6opisJR6rW9cI41epBVwQ01amJW2zeCfuSC4bO9eHQmG3birvJfEvqhddLBP9UAyGwjti4KWyfS5HGYrOGMI1T2aGvaWbAMOO-dYq22Ezmpl3PWzyb7yd1yYy2LEOqAOSuhmadQKH90cgkhBTISnC5mE8jOrwmrdZuF-Fvs4"
										alt="Avatar"
										className="w-full h-full object-cover"
									/>
								</span>
								<div className="flex-1 overflow-hidden text-left">
									<p className="text-[14px] font-bold text-slate-900 truncate">个人中心</p>
									<p className="text-[10px] text-blue-600 font-bold uppercase tracking-tight">PREMIUM</p>
								</div>
								<MoreVertical className="size-4 text-slate-400 shrink-0" />
							</button>
						}
					/>
					<DropdownMenuContent
						align="end"
						side="top"
						sideOffset={10}
						className="leros-profile-menu"
					>
						<DropdownMenuItem>
							<UserRound className="size-4" />
							<span>个人信息</span>
						</DropdownMenuItem>
						<DropdownMenuItem>
							<Settings className="size-4" />
							<span>系统设置</span>
						</DropdownMenuItem>
						<DropdownMenuItem>
							<CircleHelp className="size-4" />
							<span>使用帮助</span>
						</DropdownMenuItem>
						<DropdownMenuSeparator />
						<DropdownMenuItem variant="destructive">
							<LogOut className="size-4" />
							<span>退出登录</span>
						</DropdownMenuItem>
					</DropdownMenuContent>
				</DropdownMenu>
			</div>
		</aside>
	);
}

function ProjectList({
	projects,
	activeProjectId,
	currentView,
	onProjectClick,
}: {
	projects: Project[];
	activeProjectId: string | null;
	currentView: ViewMode;
	onProjectClick: (projectId: string) => void;
}) {
	return (
		<div className="space-y-1">
			{projects.map((project) => {
				const active = currentView === "project" && activeProjectId === project.id;
				return (
					<button
						key={project.id}
						type="button"
						onClick={() => onProjectClick(project.id)}
						className={cn(
							"flex items-center gap-3 px-3 py-1.5 w-full text-left transition-colors rounded-lg text-sm",
							active
								? "bg-slate-200/50 text-[#0050cb] font-semibold"
								: "text-slate-600 hover:text-slate-950"
						)}
					>
						<span className="text-slate-300 font-mono text-[14px]">#</span>
						<span className="truncate">{project.name}</span>
					</button>
				);
			})}
		</div>
	);
}

function NavItemButton({
	item,
	active,
	onClick,
}: {
	item: NavItem;
	active: boolean;
	onClick: () => void;
}) {
	const avatarUrl = item.icon === "IconAITeammate" ? avatarMap[item.label] : null;

	const icon = avatarUrl ? (
		<img
			src={avatarUrl}
			alt=""
			className="w-6 h-6 rounded-full object-cover flex-shrink-0"
		/>
	) : (
		iconMap[item.icon]
	);

	return (
		<button type="button" onClick={onClick} data-active={active} className="leros-nav-item">
			<span className={cn("leros-nav-icon", item.icon === "IconProject" && "leros-nav-icon-text")}>
				{icon}
			</span>
			<span className="truncate font-medium flex-1">{item.label}</span>
			{item.badge ? (
				item.icon === "IconAITeammate" ? (
					<div className="w-1.5 h-1.5 rounded-full bg-blue-600 shrink-0 mr-1" />
				) : (
					<span className="ml-auto rounded-full bg-red-100 px-1.5 py-0.5 text-xs text-red-600">
						{item.badge}
					</span>
				)
			) : null}
		</button>
	);
}
