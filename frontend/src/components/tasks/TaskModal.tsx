import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchApi } from '../../lib/api';
import { Task, TaskStatus, TaskPriority, AssigneesResponse } from '../../types';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '../ui/dialog';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Textarea } from '../ui/textarea';
import { Loader2, Trash2 } from 'lucide-react';
import { toast } from 'sonner';

const taskSchema = z.object({
  title: z.string().min(1, 'Title is required'),
  description: z.string().optional(),
  status: z.enum(['todo', 'in_progress', 'done']),
  priority: z.enum(['low', 'medium', 'high']),
  due_date: z.string().nullable().optional(),
  assignee_id: z.string().nullable().optional(),
});

type TaskFormValues = z.infer<typeof taskSchema>;

interface TaskModalProps {
  isOpen: boolean;
  onClose: () => void;
  projectId: string;
  task?: Task;
}

export function TaskModal({ isOpen, onClose, projectId, task }: TaskModalProps) {
  const queryClient = useQueryClient();
  const isEditing = !!task;

  const { register, handleSubmit, reset, setValue, watch, formState: { errors } } = useForm<TaskFormValues>({
    resolver: zodResolver(taskSchema),
    defaultValues: {
      title: '',
      description: '',
      status: 'todo',
      priority: 'medium',
      due_date: null,
      assignee_id: null,
    },
  });

  useEffect(() => {
    if (isOpen) {
      if (task) {
        reset({
          title: task.title,
          description: task.description || '',
          status: task.status,
          priority: task.priority,
          due_date: task.due_date,
          assignee_id: task.assignee_id,
        });
      } else {
        reset({
          title: '',
          description: '',
          status: 'todo',
          priority: 'medium',
          due_date: null,
          assignee_id: null,
        });
      }
    }
  }, [isOpen, task, reset]);

  const {
    data: assigneesData,
    isLoading: isAssigneesLoading,
    error: assigneesError,
  } = useQuery<AssigneesResponse>({
    queryKey: ['project-assignees', projectId],
    queryFn: () => fetchApi(`/projects/${projectId}/assignees`),
    enabled: isOpen,
  });

  const assigneeOptions = assigneesData?.users ?? [];

  const saveMutation = useMutation({
    mutationFn: (data: TaskFormValues) => {
      const payload = {
        ...data,
        assignee_id: data.assignee_id ?? null,
        due_date: data.due_date ?? null,
      };

      if (isEditing) {
        return fetchApi(`/tasks/${task.id}`, {
          method: 'PATCH',
          body: JSON.stringify(payload),
        });
      } else {
        return fetchApi(`/projects/${projectId}/tasks`, {
          method: 'POST',
          body: JSON.stringify(payload),
        });
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-detail', projectId] });
      queryClient.invalidateQueries({ queryKey: ['project-tasks', projectId] });
      toast.success(isEditing ? 'Task updated' : 'Task created');
      onClose();
    },
    onError: () => {
      toast.error(isEditing ? 'Failed to update task' : 'Failed to create task');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => fetchApi(`/tasks/${task?.id}`, { method: 'DELETE' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-detail', projectId] });
      queryClient.invalidateQueries({ queryKey: ['project-tasks', projectId] });
      toast.success('Task deleted');
      onClose();
    },
    onError: () => {
      toast.error('Failed to delete task');
    },
  });

  const onSubmit = (data: TaskFormValues) => {
    saveMutation.mutate(data);
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>{isEditing ? 'Edit Task' : 'Create Task'}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 pt-4">
          <div className="space-y-2">
            <Label htmlFor="title">Title</Label>
            <Input
              id="title"
              placeholder="What needs to be done?"
              {...register('title')}
              className={errors.title ? 'border-red-500' : ''}
            />
            {errors.title && <p className="text-sm text-red-500">{errors.title.message}</p>}
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              placeholder="Add more details..."
              {...register('description')}
              className="resize-none min-h-[100px]"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Status</Label>
              <select
                value={watch('status')}
                onChange={(e) => setValue('status', e.target.value as TaskStatus)}
                className="flex h-10 w-full rounded-md border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-stone-300"
              >
                <option value="todo">To Do</option>
                <option value="in_progress">In Progress</option>
                <option value="done">Done</option>
              </select>
            </div>
            <div className="space-y-2">
              <Label>Priority</Label>
              <select
                value={watch('priority')}
                onChange={(e) => setValue('priority', e.target.value as TaskPriority)}
                className="flex h-10 w-full rounded-md border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-stone-300"
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="due_date">Due Date</Label>
            <Input
              id="due_date"
              type="date"
              value={watch('due_date') ?? ''}
              onChange={(event) => {
                const value = event.target.value.trim();
                setValue('due_date', value === '' ? null : value, { shouldValidate: true });
              }}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="assignee_id">Assignee</Label>
            <select
              id="assignee_id"
              value={watch('assignee_id') ?? 'unassigned'}
              onChange={(event) => {
                const value = event.target.value;
                setValue('assignee_id', value === 'unassigned' ? null : value, { shouldValidate: true, shouldDirty: true });
              }}
              className="flex h-10 w-full rounded-md border border-stone-200 bg-white px-3 py-2 text-sm text-stone-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-stone-300"
              disabled={isAssigneesLoading}
            >
              <option value="unassigned">Unassigned</option>
              {assigneeOptions.map((user) => (
                <option key={user.id} value={user.id}>
                  {user.name} ({user.email})
                </option>
              ))}
            </select>
            {assigneesError ? (
              <p className="text-sm text-amber-600">Could not load assignee options. You can still save as Unassigned.</p>
            ) : !isAssigneesLoading && assigneeOptions.length === 0 ? (
              <p className="text-xs text-stone-500">No assignee options found for this project yet. New tasks can remain Unassigned.</p>
            ) : (
              <p className="text-xs text-stone-500">Choose a team member or keep it unassigned.</p>
            )}
          </div>

          <DialogFooter className="pt-4 flex justify-between sm:justify-between">
            {isEditing ? (
              <Button 
                type="button" 
                variant="destructive" 
                size="icon"
                onClick={() => {
                  if (confirm('Are you sure you want to delete this task?')) {
                    deleteMutation.mutate();
                  }
                }}
                disabled={deleteMutation.isPending}
              >
                {deleteMutation.isPending ? <Loader2 className="w-4 h-4 animate-spin" /> : <Trash2 className="w-4 h-4" />}
              </Button>
            ) : (
              <div></div>
            )}
            <div className="flex gap-2">
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
              <Button type="submit" disabled={saveMutation.isPending} className="bg-stone-900 text-white">
                {saveMutation.isPending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                {isEditing ? 'Save Changes' : 'Create Task'}
              </Button>
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
