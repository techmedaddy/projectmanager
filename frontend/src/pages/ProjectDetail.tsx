import { useMemo, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ApiError, fetchApi } from '../lib/api';
import { ProjectDetailResponse, ProjectStatsResponse, Task, TaskStatus, TasksResponse } from '../types';
import { Button } from '../components/ui/button';
import { Loader2, ArrowLeft, Plus } from 'lucide-react';
import { TaskBoard } from '../components/tasks/TaskBoard';
import { TaskModal } from '../components/tasks/TaskModal';

export function ProjectDetail() {
  const { id } = useParams<{ id: string }>();
  const [isTaskModalOpen, setIsTaskModalOpen] = useState(false);
  const [selectedTask, setSelectedTask] = useState<Task | undefined>();
  const [statusFilter, setStatusFilter] = useState<'all' | TaskStatus>('all');
  const [assigneeFilter, setAssigneeFilter] = useState<'all' | string>('all');
  const [tasksPage, setTasksPage] = useState(1);
  const tasksLimit = 9;

  const { data: projectData, isLoading: isProjectLoading, error: projectError } = useQuery<ProjectDetailResponse>({
    queryKey: ['project-detail', id],
    queryFn: () => fetchApi(`/projects/${id}`),
    enabled: !!id,
  });

  const taskFilters = useMemo(() => {
    const params = new URLSearchParams();

    params.set('page', String(tasksPage));
    params.set('limit', String(tasksLimit));

    if (statusFilter !== 'all') {
      params.set('status', statusFilter);
    }

    if (assigneeFilter !== 'all') {
      params.set('assignee', assigneeFilter);
    }

    const queryString = params.toString();
    return queryString ? `?${queryString}` : '';
  }, [statusFilter, assigneeFilter, tasksPage]);

  const {
    data: tasksData,
    isLoading: isTasksLoading,
    isFetching: isTasksFetching,
    error: tasksError,
    refetch: refetchTasks,
  } = useQuery<TasksResponse>({
    queryKey: ['project-tasks', id, statusFilter, assigneeFilter, tasksPage, tasksLimit],
    queryFn: () => fetchApi(`/projects/${id}/tasks${taskFilters}`),
    enabled: !!id,
    placeholderData: (previousData) => previousData,
  });

  const { data: statsData } = useQuery<ProjectStatsResponse>({
    queryKey: ['project-stats', id],
    queryFn: () => fetchApi(`/projects/${id}/stats`),
    enabled: !!id,
  });

  const { data: assigneesData } = useQuery<{ users: { id: string; name: string; email: string }[] }>({
    queryKey: ['project-assignees', id],
    queryFn: () => fetchApi(`/projects/${id}/assignees`),
    enabled: !!id,
  });

  const allProjectTasks = projectData?.tasks ?? [];
  const tasks = tasksData?.items ?? tasksData?.tasks ?? [];
  const tasksMeta = tasksData?.meta;

  const assigneeOptions = useMemo(() => {
    return Array.from(
      new Set(allProjectTasks.map((task) => task.assignee_id).filter((value): value is string => value !== null))
    );
  }, [allProjectTasks]);

  const hasActiveFilters = statusFilter !== 'all' || assigneeFilter !== 'all';

  const tasksTotalPages = useMemo(() => {
    const total = tasksMeta?.total ?? tasks.length;
    const limit = tasksMeta?.limit ?? tasksLimit;
    return Math.max(1, Math.ceil(total / limit));
  }, [tasksMeta, tasks.length]);

  const assigneeNameByID = useMemo(() => {
    const map = new Map<string, string>();
    (assigneesData?.users ?? []).forEach((user) => {
      map.set(user.id, `${user.name} (${user.email})`);
    });
    return map;
  }, [assigneesData]);

  const topAssignees = useMemo(() => {
    if (!statsData?.by_assignee) return [] as Array<{ key: string; count: number }>;

    return Object.entries(statsData.by_assignee)
      .sort((a, b) => b[1] - a[1])
      .slice(0, 3)
      .map(([key, count]) => ({ key, count }));
  }, [statsData]);

  const handleOpenCreate = () => {
    setSelectedTask(undefined);
    setIsTaskModalOpen(true);
  };

  const handleOpenEdit = (task: Task) => {
    setSelectedTask(task);
    setIsTaskModalOpen(true);
  };

  if (isProjectLoading) {
    return (
      <div className="flex justify-center py-16">
        <Loader2 className="w-8 h-8 animate-spin text-cyan-300" />
      </div>
    );
  }

  if (projectError || !projectData) {
    const apiError = projectError instanceof ApiError ? projectError : null;
    const status = apiError?.status;
    const message =
      status === 401
        ? 'Your session expired. Please sign in again.'
        : status === 403
          ? 'You do not have access to this project.'
          : status === 404
            ? 'Project not found.'
            : 'Failed to load project details.';

    return (
      <div className="text-center py-12 text-slate-200">
        <p className="text-rose-300">{message}</p>
        {status && <p className="text-xs text-slate-400 mt-1">Error code: {status}</p>}
        <Link to="/projects" className="text-slate-300 hover:text-cyan-300 mt-4 inline-block">
          &larr; Back to Projects
        </Link>
      </div>
    );
  }

  const { project } = projectData;

  const handleResetFilters = () => {
    setStatusFilter('all');
    setAssigneeFilter('all');
    setTasksPage(1);
  };

  return (
    <div className="space-y-8 h-full flex flex-col text-slate-100">
      <section className="rounded-3xl border border-white/10 bg-gradient-to-r from-indigo-500/20 via-cyan-500/10 to-transparent p-6 shadow-[0_20px_50px_rgba(0,0,0,0.25)]">
        <div className="flex items-start justify-between gap-4">
          <div>
            <Link to="/projects" className="inline-flex items-center text-sm text-slate-300 hover:text-cyan-300 mb-4 transition-colors">
              <ArrowLeft className="w-4 h-4 mr-1" />
              Projects
            </Link>
            <h1 className="text-3xl font-semibold tracking-tight text-white leading-tight">{project.name}</h1>
            {project.description && <p className="text-slate-300 mt-2 max-w-2xl">{project.description}</p>}
          </div>
          <div className="flex items-center gap-3">
            <Button onClick={handleOpenCreate} className="bg-cyan-400 text-slate-950 hover:bg-cyan-300 shadow-lg shadow-cyan-500/30">
              <Plus className="w-4 h-4 mr-2" />
              Add Task
            </Button>
          </div>
        </div>
      </section>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="rounded-2xl border border-white/10 bg-slate-900/55 p-4 shadow-sm backdrop-blur-sm">
          <p className="text-xs font-medium uppercase tracking-wide text-slate-300">Status totals</p>
          <div className="mt-2 grid grid-cols-3 gap-2 text-sm">
            <div className="rounded-md bg-slate-800/70 p-2 text-center border border-white/5">
              <p className="text-slate-300">To Do</p>
              <p className="font-semibold text-white">{statsData?.by_status?.todo ?? 0}</p>
            </div>
            <div className="rounded-md bg-slate-800/70 p-2 text-center border border-white/5">
              <p className="text-slate-300">In Progress</p>
              <p className="font-semibold text-white">{statsData?.by_status?.in_progress ?? 0}</p>
            </div>
            <div className="rounded-md bg-slate-800/70 p-2 text-center border border-white/5">
              <p className="text-slate-300">Done</p>
              <p className="font-semibold text-white">{statsData?.by_status?.done ?? 0}</p>
            </div>
          </div>
        </div>

        <div className="rounded-2xl border border-white/10 bg-slate-900/55 p-4 shadow-sm backdrop-blur-sm">
          <p className="text-xs font-medium uppercase tracking-wide text-slate-300">Top assignees</p>
          <div className="mt-2 space-y-1 text-sm">
            {topAssignees.length === 0 ? (
              <p className="text-slate-300">No assignee activity yet.</p>
            ) : (
              topAssignees.map((item) => (
                <div key={item.key} className="flex items-center justify-between rounded-md bg-slate-800/70 border border-white/5 px-2 py-1">
                  <span className="truncate text-slate-200">
                    {item.key === 'unassigned' ? 'Unassigned' : (assigneeNameByID.get(item.key) ?? item.key)}
                  </span>
                  <span className="font-medium text-white">{item.count}</span>
                </div>
              ))
            )}
          </div>
        </div>
      </div>

      <div className="flex flex-col gap-3 rounded-2xl border border-white/10 bg-slate-900/50 p-4 shadow-sm backdrop-blur-sm sm:flex-row sm:items-end sm:justify-between">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
          <label className="space-y-1">
            <span className="text-xs font-medium uppercase tracking-wide text-slate-300">Status</span>
            <select
              value={statusFilter}
              onChange={(event) => {
                setStatusFilter(event.target.value as 'all' | TaskStatus);
                setTasksPage(1);
              }}
              className="h-10 w-full rounded-md border border-white/15 bg-slate-800/80 px-3 text-sm text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300/50 sm:min-w-44"
            >
              <option value="all">All statuses</option>
              <option value="todo">To Do</option>
              <option value="in_progress">In Progress</option>
              <option value="done">Done</option>
            </select>
          </label>

          <label className="space-y-1">
            <span className="text-xs font-medium uppercase tracking-wide text-slate-300">Assignee</span>
            <select
              value={assigneeFilter}
              onChange={(event) => {
                setAssigneeFilter(event.target.value);
                setTasksPage(1);
              }}
              className="h-10 w-full rounded-md border border-white/15 bg-slate-800/80 px-3 text-sm text-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300/50 sm:min-w-56"
            >
              <option value="all">All assignees</option>
              {assigneeOptions.map((assigneeID) => (
                <option key={assigneeID} value={assigneeID}>
                  {assigneeID}
                </option>
              ))}
            </select>
          </label>
        </div>

        <div className="flex items-center justify-between gap-3 sm:justify-end">
          <p className="text-sm text-slate-300">
            Showing <span className="font-medium text-white">{tasks.length}</span>
            {hasActiveFilters ? ' filtered' : ''} tasks
            {isTasksFetching ? ' · Updating…' : ''}
          </p>
          <Button variant="outline" onClick={handleResetFilters} disabled={!hasActiveFilters}>
            Clear filters
          </Button>
        </div>
      </div>

      <div className="flex-1 min-h-0 overflow-x-auto pb-4">
        {isTasksLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-7 w-7 animate-spin text-cyan-300" />
          </div>
        ) : tasksError ? (
          (() => {
            const apiError = tasksError instanceof ApiError ? tasksError : null;
            const status = apiError?.status;
            const message =
              status === 401
                ? 'Your session expired. Please sign in again.'
                : status === 403
                  ? 'You do not have access to these tasks.'
                  : status === 404
                    ? 'Tasks resource not found for this project.'
                    : 'Failed to load filtered tasks.';

            return (
              <div className="rounded-xl border border-rose-300/30 bg-rose-500/10 p-4 text-sm text-rose-200">
                <p>{message}</p>
                {status && <p className="text-xs text-rose-200/80 mt-1">Error code: {status}</p>}
                <Button variant="outline" className="mt-3" onClick={() => void refetchTasks()}>
                  Retry
                </Button>
              </div>
            );
          })()
        ) : hasActiveFilters && tasks.length === 0 ? (
          <div className="rounded-xl border border-dashed border-white/20 bg-slate-900/45 p-8 text-center">
            <p className="text-base font-medium text-white">No tasks match current filters</p>
            <p className="mt-1 text-sm text-slate-300">Try another status/assignee combination or clear filters.</p>
            <Button variant="outline" className="mt-4" onClick={handleResetFilters}>
              Clear filters
            </Button>
          </div>
        ) : (
          <>
            <TaskBoard tasks={tasks} projectId={project.id} onTaskClick={handleOpenEdit} />

            <div className="mt-4 flex items-center justify-between rounded-2xl border border-white/10 bg-slate-900/45 p-3.5 shadow-sm backdrop-blur-sm">
              <p className="text-sm text-slate-300">
                Page <span className="font-medium text-white">{tasksMeta?.page ?? tasksPage}</span> · Total tasks{' '}
                <span className="font-medium text-white">{tasksMeta?.total ?? tasks.length}</span>
                {isTasksFetching ? ' · Updating…' : ''}
              </p>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  onClick={() => setTasksPage((p) => Math.max(1, p - 1))}
                  disabled={(tasksMeta?.page ?? tasksPage) <= 1 || isTasksFetching}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setTasksPage((p) => Math.min(tasksTotalPages, p + 1))}
                  disabled={(tasksMeta?.page ?? tasksPage) >= tasksTotalPages || isTasksFetching}
                >
                  Next
                </Button>
              </div>
            </div>
          </>
        )}
      </div>

      <TaskModal
        isOpen={isTaskModalOpen}
        onClose={() => setIsTaskModalOpen(false)}
        projectId={project.id}
        task={selectedTask}
      />
    </div>
  );
}
