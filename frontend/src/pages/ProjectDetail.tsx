import { useMemo, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchApi } from '../lib/api';
import { ProjectDetailResponse, Task, TaskStatus, TasksResponse } from '../types';
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

  const { data: projectData, isLoading: isProjectLoading, error: projectError } = useQuery<ProjectDetailResponse>({
    queryKey: ['project-detail', id],
    queryFn: () => fetchApi(`/projects/${id}`),
    enabled: !!id,
  });

  const taskFilters = useMemo(() => {
    const params = new URLSearchParams();

    if (statusFilter !== 'all') {
      params.set('status', statusFilter);
    }

    if (assigneeFilter !== 'all') {
      params.set('assignee', assigneeFilter);
    }

    const queryString = params.toString();
    return queryString ? `?${queryString}` : '';
  }, [statusFilter, assigneeFilter]);

  const {
    data: tasksData,
    isLoading: isTasksLoading,
    isFetching: isTasksFetching,
    error: tasksError,
    refetch: refetchTasks,
  } = useQuery<TasksResponse>({
    queryKey: ['project-tasks', id, statusFilter, assigneeFilter],
    queryFn: () => fetchApi(`/projects/${id}/tasks${taskFilters}`),
    enabled: !!id,
    placeholderData: (previousData) => previousData,
  });

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
      <div className="flex justify-center py-12">
        <Loader2 className="w-8 h-8 animate-spin text-stone-400" />
      </div>
    );
  }

  if (projectError || !projectData) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500">Failed to load project details.</p>
        <Link to="/projects" className="text-stone-500 hover:text-stone-900 mt-4 inline-block">
          &larr; Back to Projects
        </Link>
      </div>
    );
  }

  const { project, tasks: allProjectTasks } = projectData;
  const tasks = tasksData?.tasks ?? [];

  const assigneeOptions = useMemo(() => {
    return Array.from(
      new Set(allProjectTasks.map((task) => task.assignee_id).filter((value): value is string => value !== null))
    );
  }, [allProjectTasks]);

  const hasActiveFilters = statusFilter !== 'all' || assigneeFilter !== 'all';

  const handleResetFilters = () => {
    setStatusFilter('all');
    setAssigneeFilter('all');
  };

  return (
    <div className="space-y-8 h-full flex flex-col">
      <div className="flex items-start justify-between">
        <div>
          <Link to="/projects" className="inline-flex items-center text-sm text-stone-500 hover:text-stone-900 mb-4 transition-colors">
            <ArrowLeft className="w-4 h-4 mr-1" />
            Projects
          </Link>
          <h1 className="text-3xl font-semibold tracking-tight text-stone-900">{project.name}</h1>
          {project.description && (
            <p className="text-stone-500 mt-2 max-w-2xl">{project.description}</p>
          )}
        </div>
        <div className="flex items-center gap-3">
          <Button onClick={handleOpenCreate} className="bg-stone-900 hover:bg-stone-800 text-white">
            <Plus className="w-4 h-4 mr-2" />
            Add Task
          </Button>
        </div>
      </div>

      <div className="flex flex-col gap-3 rounded-xl border border-stone-200 bg-white p-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 sm:gap-4">
          <label className="space-y-1">
            <span className="text-xs font-medium uppercase tracking-wide text-stone-500">Status</span>
            <select
              value={statusFilter}
              onChange={(event) => setStatusFilter(event.target.value as 'all' | TaskStatus)}
              className="h-10 w-full rounded-md border border-stone-200 bg-white px-3 text-sm text-stone-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-stone-300 sm:min-w-44"
            >
              <option value="all">All statuses</option>
              <option value="todo">To Do</option>
              <option value="in_progress">In Progress</option>
              <option value="done">Done</option>
            </select>
          </label>

          <label className="space-y-1">
            <span className="text-xs font-medium uppercase tracking-wide text-stone-500">Assignee</span>
            <select
              value={assigneeFilter}
              onChange={(event) => setAssigneeFilter(event.target.value)}
              className="h-10 w-full rounded-md border border-stone-200 bg-white px-3 text-sm text-stone-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-stone-300 sm:min-w-56"
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
          <p className="text-sm text-stone-500">
            Showing <span className="font-medium text-stone-700">{tasks.length}</span>
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
            <Loader2 className="h-7 w-7 animate-spin text-stone-400" />
          </div>
        ) : tasksError ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
            <p>Failed to load filtered tasks.</p>
            <Button variant="outline" className="mt-3" onClick={() => void refetchTasks()}>
              Retry
            </Button>
          </div>
        ) : hasActiveFilters && tasks.length === 0 ? (
          <div className="rounded-xl border border-dashed border-stone-300 bg-white p-8 text-center">
            <p className="text-base font-medium text-stone-800">No tasks match current filters</p>
            <p className="mt-1 text-sm text-stone-500">Try another status/assignee combination or clear filters.</p>
            <Button variant="outline" className="mt-4" onClick={handleResetFilters}>
              Clear filters
            </Button>
          </div>
        ) : (
          <TaskBoard tasks={tasks} projectId={project.id} onTaskClick={handleOpenEdit} />
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
