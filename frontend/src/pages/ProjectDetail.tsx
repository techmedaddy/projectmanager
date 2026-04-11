import { useMemo, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchApi } from '../lib/api';
import { ProjectDetailResponse, Task, TaskStatus } from '../types';
import { Button } from '../components/ui/button';
import { Loader2, ArrowLeft, Plus } from 'lucide-react';
import { TaskBoard } from '../components/tasks/TaskBoard';
import { TaskModal } from '../components/tasks/TaskModal';

export function ProjectDetail() {
  const { id } = useParams<{ id: string }>();
  const [isTaskModalOpen, setIsTaskModalOpen] = useState(false);
  const [selectedTask, setSelectedTask] = useState<Task | undefined>();
  const [statusFilter, setStatusFilter] = useState<'all' | TaskStatus>('all');
  const [assigneeFilter, setAssigneeFilter] = useState<'all' | 'unassigned' | string>('all');

  const { data, isLoading, error } = useQuery<ProjectDetailResponse>({
    queryKey: ['project', id],
    queryFn: () => fetchApi(`/projects/${id}`),
    enabled: !!id,
  });

  const handleOpenCreate = () => {
    setSelectedTask(undefined);
    setIsTaskModalOpen(true);
  };

  const handleOpenEdit = (task: Task) => {
    setSelectedTask(task);
    setIsTaskModalOpen(true);
  };

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <Loader2 className="w-8 h-8 animate-spin text-stone-400" />
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500">Failed to load project details.</p>
        <Link to="/projects" className="text-stone-500 hover:text-stone-900 mt-4 inline-block">
          &larr; Back to Projects
        </Link>
      </div>
    );
  }

  const { project, tasks } = data;

  const assigneeOptions = useMemo(() => {
    return Array.from(
      new Set(tasks.map((task) => task.assignee_id).filter((value): value is string => value !== null))
    );
  }, [tasks]);

  const filteredTasks = useMemo(() => {
    return tasks.filter((task) => {
      const statusMatches = statusFilter === 'all' || task.status === statusFilter;
      const assigneeMatches =
        assigneeFilter === 'all' ||
        (assigneeFilter === 'unassigned' ? task.assignee_id === null : task.assignee_id === assigneeFilter);

      return statusMatches && assigneeMatches;
    });
  }, [tasks, statusFilter, assigneeFilter]);

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
              className="h-10 min-w-44 rounded-md border border-stone-200 bg-white px-3 text-sm text-stone-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-stone-300"
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
              className="h-10 min-w-56 rounded-md border border-stone-200 bg-white px-3 text-sm text-stone-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-stone-300"
            >
              <option value="all">All assignees</option>
              <option value="unassigned">Unassigned</option>
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
            Showing <span className="font-medium text-stone-700">{filteredTasks.length}</span> of {tasks.length} tasks
          </p>
          <Button variant="outline" onClick={handleResetFilters} disabled={!hasActiveFilters}>
            Clear filters
          </Button>
        </div>
      </div>

      <div className="flex-1 min-h-0 overflow-x-auto pb-4">
        <TaskBoard 
          tasks={filteredTasks} 
          projectId={project.id} 
          onTaskClick={handleOpenEdit} 
        />
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
