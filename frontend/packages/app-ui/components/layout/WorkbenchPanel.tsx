"use client";

import { useLayoutStore } from "@leros/store";
import { Button } from "@leros/ui/components/ui/button";
import { Bell, ChevronDown, Folder, Plus, Search, SendHorizonal } from "lucide-react";
import { useState } from "react";

const mockActivities = [
	{
		id: "activity-1",
		avatar: "SK",
		name: "Sarah K.",
		project: "backend-v2",
		time: "2 分钟前",
		description: "完成了 API 追踪",
		note: "解决了 auth-middleware 模块中的 4 个延迟问题。系统开销降低了 12%。",
	},
	{
		id: "activity-2",
		avatar: "AL",
		name: "Ada Lovelace",
		project: "frontend-core",
		time: "45 分钟前",
		description: "更新了文档",
		tags: ["文档", "修订版本 3"],
	},
];

export function WorkbenchPanel() {
	const { projects, activeProjectId, selectWorkbenchProject, sendWorkbenchMessage, switchProject } =
		useLayoutStore((s) => s);
	const [input, setInput] = useState("");

	const handleSend = () => {
		if (!input.trim()) return;
		sendWorkbenchMessage(input, activeProjectId);
		setInput("");
	};
	const activeProject = projects.find((project) => project.id === activeProjectId);
	const latestProject = projects[0];

	return (
		<div data-slot="workbench-panel" className="min-h-0 flex-1 overflow-y-auto bg-[#f7f8fd]">
			{/* Top Header */}
			<header className="h-16 flex justify-end items-center px-10 z-10 shrink-0">
				<div className="flex items-center gap-4 text-slate-700">
					<button type="button" className="p-2 hover:bg-slate-200/50 rounded-full transition-colors">
						<Search className="size-5" />
					</button>
					<button type="button" className="p-2 hover:bg-slate-200/50 rounded-full transition-colors relative">
						<Bell className="size-5" />
						<span className="absolute top-2 right-2 size-2 bg-red-500 rounded-full border-2 border-[#f7f8fd]" />
					</button>
				</div>
			</header>

			{/* Main Content Canvas */}
			<div className="max-w-[1100px] mx-auto w-full px-10 py-12 flex-1 flex flex-col z-10">
				{/* Welcome/Hero Section */}
				<section className="mb-8">
					<div className="max-w-3xl mx-auto">
						<div className="flex flex-col gap-4 items-start text-left mb-6">
							<h2 className="font-bold text-slate-900 tracking-tight text-4xl md:text-5xl">
								Hi, <span className="text-blue-600">Mia</span>
							</h2>
							<p className="text-lg font-medium text-slate-400 tracking-widest uppercase">
								以 Leros 智能赋能您的工作流。
							</p>
						</div>

						{/* Enhanced Command Input Card */}
						<div className="bg-white border border-slate-200/60 rounded-[24px] shadow-sm flex flex-col p-4 focus-within:border-blue-600/50 focus-within:shadow-md transition-all">
							<div className="flex gap-3 mb-2">
								<textarea
									value={input}
									onChange={(event) => setInput(event.target.value)}
									onKeyDown={(event) => {
										if (event.key === "Enter" && !event.shiftKey) {
											event.preventDefault();
											handleSend();
										}
									}}
									placeholder="在这里开始新任务，或输入指令以同步您的项目进度..."
									className="flex-1 bg-transparent border-none focus:ring-0 text-base text-slate-700 placeholder:text-slate-300 resize-none h-[60px] outline-none"
								/>
							</div>
							<div className="flex items-center justify-between border-t border-slate-100 pt-3">
								<div className="flex items-center gap-3">
									<button
										type="button"
										className="p-1.5 text-slate-500 hover:bg-slate-100 rounded-full transition-colors"
										aria-label="添加附件"
									>
										<Plus className="size-5" />
									</button>
									<div className="relative">
										<Folder className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-slate-500" />
										<select
											value={activeProjectId ?? ""}
											onChange={(event) => selectWorkbenchProject(event.target.value || null)}
											className="h-8 min-w-[140px] appearance-none rounded-full border border-slate-200 bg-white pl-9 pr-8 text-xs font-semibold text-slate-700 outline-none transition-colors hover:border-slate-300"
											aria-label="新项目"
										>
											<option value="">选择项目</option>
											{projects.map((project) => (
												<option key={project.id} value={project.id}>
													{project.name}
												</option>
											))}
										</select>
										<ChevronDown className="pointer-events-none absolute right-3 top-1/2 size-3.5 -translate-y-1/2 text-slate-400" />
									</div>
									{activeProject && (
										<button
											type="button"
											onClick={() => switchProject(activeProject.id)}
											className="text-xs font-semibold text-blue-600 hover:underline"
										>
											打开项目
										</button>
									)}
								</div>
								<div className="flex items-center gap-2">
									<Button
										size="icon"
										onClick={handleSend}
										disabled={!input.trim()}
										className="size-9 rounded-xl bg-blue-600 text-white shadow-sm hover:bg-blue-700 disabled:bg-slate-100 disabled:text-slate-300"
									>
										<SendHorizonal className="size-4" />
									</Button>
								</div>
							</div>
						</div>

						{/* Suggested Actions */}
						<div className="flex items-center gap-4 mt-6 justify-center">
							<button className="text-[11px] font-bold text-slate-400 hover:text-blue-600 transition-colors flex items-center gap-1.5 uppercase tracking-widest">
								分析 SPRINT
							</button>
							<button className="text-[11px] font-bold text-slate-400 hover:text-blue-600 transition-colors flex items-center gap-1.5 uppercase tracking-widest">
								总结报告
							</button>
						</div>
					</div>
				</section>

				{/* Workbench Grid */}
				<section className="grid grid-cols-12 gap-10 flex-1 mt-6">
					{/* Left: Activity Stream (col-span-8) */}
					<div className="col-span-8">
						<div className="flex items-center justify-between mb-8 border-b border-slate-200/50 pb-4">
							<h3 className="text-xl font-bold text-slate-900 tracking-tight">动态流</h3>
							<div className="flex bg-slate-200/50 p-1 rounded-lg">
								<button type="button" className="px-4 py-1.5 text-[12px] font-bold bg-white text-slate-900 rounded shadow-sm">今日</button>
								<button type="button" className="px-4 py-1.5 text-[12px] font-bold text-slate-500 hover:text-slate-900">本周</button>
							</div>
						</div>

						<div className="space-y-10 relative">
							{mockActivities.map((activity, idx) => (
								<div key={activity.id} className="flex gap-6 relative">
									{/* Vertical timeline line */}
									{idx < mockActivities.length - 1 && (
										<div className="absolute left-[19px] top-10 bottom-[-40px] w-[1px] bg-slate-200/80" />
									)}
									<div className="flex-shrink-0 z-10">
										<div className="flex size-10 items-center justify-center rounded-full bg-slate-900 text-sm font-bold text-white border-2 border-white shadow-sm">
											{activity.avatar}
										</div>
									</div>
									<div className="flex-1 pt-0.5">
										<div className="flex items-baseline justify-between mb-2">
											<p className="text-sm text-slate-700">
												<span className="font-bold text-slate-900">{activity.name}</span>
												<span> 在 </span>
												<span className="text-blue-600 font-semibold hover:underline cursor-pointer" onClick={() => switchProject(activity.project)}>
													{activity.project}
												</span>
												<span> 中{activity.description}</span>
											</p>
											<span className="text-[11px] text-slate-400 font-medium">{activity.time}</span>
										</div>
										{activity.note ? (
											<div className="p-4 bg-slate-100/70 rounded-xl text-[13px] text-slate-600 leading-relaxed">
												“{activity.note}”
											</div>
										) : null}
										{activity.tags ? (
											<div className="flex items-center gap-2 mt-2">
												{activity.tags.map((tag) => (
													<span key={tag} className="px-2.5 py-1 bg-slate-200/50 text-slate-600 text-[10px] rounded font-bold tracking-wider uppercase">
														{tag}
													</span>
												))}
											</div>
										) : null}
									</div>
								</div>
							))}

							{latestProject && (
								<div className="rounded-2xl border border-blue-100 bg-blue-50/60 px-5 py-4 text-xs text-blue-900 mt-6">
									最近项目：{latestProject.name} · {latestProject.description}
								</div>
							)}
						</div>
					</div>

					{/* Right: Stats & Promotion (col-span-4) */}
					<div className="col-span-4 space-y-10">
						{/* ToDo card */}
						<div className="p-6 border border-slate-200/50 rounded-2xl bg-white shadow-sm">
							<h4 className="text-[11px] font-bold text-slate-400 uppercase tracking-widest mb-6">待办事项</h4>
							<div className="space-y-5">
								<div className="flex flex-col gap-2 p-2 -mx-2 hover:bg-slate-50 rounded-lg transition-colors cursor-pointer group">
									<div className="flex items-center justify-between">
										<p className="text-[13px] font-bold text-slate-900 flex-1 truncate">优化数据库查询性能</p>
										<span className="px-2 py-0.5 bg-red-100 text-red-600 text-[10px] rounded font-bold uppercase tracking-wider">待处理</span>
									</div>
									<div className="flex items-center gap-2 text-slate-400">
										<span className="text-[11px] font-medium">backend-v2</span>
									</div>
								</div>
								<div className="flex flex-col gap-2 p-2 -mx-2 hover:bg-slate-50 rounded-lg transition-colors cursor-pointer group">
									<div className="flex items-center justify-between">
										<p className="text-[13px] font-bold text-slate-900 flex-1 truncate">前端 UI 组件化重构</p>
										<span className="px-2 py-0.5 bg-blue-100 text-blue-600 text-[10px] rounded font-bold uppercase tracking-wider">进行中</span>
									</div>
									<div className="flex items-center gap-2 text-slate-400">
										<span className="text-[11px] font-medium">frontend-core</span>
									</div>
								</div>
								<div className="flex flex-col gap-2 p-2 -mx-2 hover:bg-slate-50 rounded-lg transition-colors cursor-pointer group">
									<div className="flex items-center justify-between">
										<p className="text-[13px] font-bold text-slate-900 flex-1 truncate">基础设施安全审计</p>
										<span className="px-2 py-0.5 bg-slate-100 text-slate-600 text-[10px] rounded font-bold uppercase tracking-wider">待处理</span>
									</div>
									<div className="flex items-center gap-2 text-slate-400">
										<span className="text-[11px] font-medium">infra</span>
									</div>
								</div>
							</div>
						</div>

						{/* Recent Visit files card */}
						<div className="p-6 border border-slate-200/50 rounded-2xl bg-white shadow-sm">
							<h4 className="text-[11px] font-bold text-slate-400 uppercase tracking-widest mb-6">最近访问</h4>
							<div className="space-y-5">
								<div className="flex items-center gap-4 hover:bg-slate-50 p-2 -mx-2 rounded-lg transition-colors cursor-pointer">
									<div className="w-9 h-9 flex items-center justify-center bg-red-50 text-red-500 rounded-lg">
										<span className="text-xs font-bold font-mono">PDF</span>
									</div>
									<div className="flex-1 min-w-0">
										<p className="text-[13px] font-bold text-slate-900 truncate">Q4 产品规划指南.pdf</p>
										<p className="text-[11px] text-slate-400">今天 10:45</p>
									</div>
								</div>
								<div className="flex items-center gap-4 hover:bg-slate-50 p-2 -mx-2 rounded-lg transition-colors cursor-pointer">
									<div className="w-9 h-9 flex items-center justify-center bg-blue-50 text-blue-500 rounded-lg">
										<span className="text-xs font-bold font-mono">DOC</span>
									</div>
									<div className="flex-1 min-w-0">
										<p className="text-[13px] font-bold text-slate-900 truncate">后端架构重构草案.docx</p>
										<p className="text-[11px] text-slate-400">昨天 16:20</p>
									</div>
								</div>
								<div className="flex items-center gap-4 hover:bg-slate-50 p-2 -mx-2 rounded-lg transition-colors cursor-pointer">
									<div className="w-9 h-9 flex items-center justify-center bg-slate-100 text-slate-500 rounded-lg">
										<span className="text-xs font-bold font-mono">PNG</span>
									</div>
									<div className="flex-1 min-w-0">
										<p className="text-[13px] font-bold text-slate-900 truncate">v0.2 设计手稿.png</p>
										<p className="text-[11px] text-slate-400">10月24日</p>
									</div>
								</div>
							</div>
						</div>
					</div>
				</section>
			</div>
		</div>
	);
}
