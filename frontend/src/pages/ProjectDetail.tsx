import { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchApi } from '../lib/api';
import { ProjectDetailResponse, Task, TaskStatus } from '../types';
import { Button } from '../components/ui/button';
import { Loader2, ArrowLeft, Plus, MoreHorizontal } from 'lucide-react';
import { TaskBoard } from '../components/tasks/TaskBoard';
import { TaskModal } from '../components/tasks/TaskModal';
import { toast } from 'sonner';

export function ProjectDetail() {
  const { id } = useParams<{ id: string }>();
  const queryClient = useQueryClient();
  const [isTaskModalOpen, setIsTaskModalOpen] = useState(false);
  const [selectedTask, setSelectedTask] = useState<Task | undefined>();

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

      <div className="flex-1 min-h-0 overflow-x-auto pb-4">
        <TaskBoard 
          tasks={tasks} 
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
