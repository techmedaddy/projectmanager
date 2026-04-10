import React from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Task, TaskStatus } from '../../types';
import { fetchApi } from '../../lib/api';
import { toast } from 'sonner';
import { format } from 'date-fns';
import { Calendar, Clock, CheckCircle2, Circle, ArrowRightCircle } from 'lucide-react';
import { Badge } from '../ui/badge';
import { cn } from '../../lib/utils';

interface TaskBoardProps {
  tasks: Task[];
  projectId: string;
  onTaskClick: (task: Task) => void;
}

const COLUMNS: { id: TaskStatus; title: string; icon: React.ElementType }[] = [
  { id: 'todo', title: 'To Do', icon: Circle },
  { id: 'in_progress', title: 'In Progress', icon: ArrowRightCircle },
  { id: 'done', title: 'Done', icon: CheckCircle2 },
];

export function TaskBoard({ tasks, projectId, onTaskClick }: TaskBoardProps) {
  const queryClient = useQueryClient();

  const updateStatusMutation = useMutation({
    mutationFn: ({ taskId, status }: { taskId: string; status: TaskStatus }) =>
      fetchApi(`/tasks/${taskId}`, {
        method: 'PATCH',
        body: JSON.stringify({ status }),
      }),
    onMutate: async ({ taskId, status }) => {
      await queryClient.cancelQueries({ queryKey: ['project', projectId] });
      const previousData = queryClient.getQueryData(['project', projectId]);

      queryClient.setQueryData(['project', projectId], (old: any) => {
        if (!old) return old;
        return {
          ...old,
          tasks: old.tasks.map((t: Task) =>
            t.id === taskId ? { ...t, status } : t
          ),
        };
      });

      return { previousData };
    },
    onError: (err, newTodo, context) => {
      queryClient.setQueryData(['project', projectId], context?.previousData);
      toast.error('Failed to update task status');
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
    },
  });

  const handleStatusChange = (e: React.MouseEvent, taskId: string, newStatus: TaskStatus) => {
    e.stopPropagation();
    updateStatusMutation.mutate({ taskId, status: newStatus });
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'bg-red-100 text-red-700 border-red-200';
      case 'medium': return 'bg-amber-100 text-amber-700 border-amber-200';
      case 'low': return 'bg-emerald-100 text-emerald-700 border-emerald-200';
      default: return 'bg-stone-100 text-stone-700 border-stone-200';
    }
  };

  return (
    <div className="flex gap-6 h-full items-start">
      {COLUMNS.map((column) => {
        const columnTasks = tasks.filter((t) => t.status === column.id);
        const Icon = column.icon;

        return (
          <div key={column.id} className="flex-1 min-w-[300px] bg-stone-100/50 rounded-2xl p-4 border border-stone-200/60">
            <div className="flex items-center justify-between mb-4 px-1">
              <div className="flex items-center gap-2 text-stone-700 font-medium">
                <Icon className={cn("w-4 h-4", 
                  column.id === 'todo' ? 'text-stone-400' : 
                  column.id === 'in_progress' ? 'text-orange-500' : 
                  'text-emerald-500'
                )} />
                {column.title}
              </div>
              <Badge variant="secondary" className="bg-stone-200/50 text-stone-600 hover:bg-stone-200/50">
                {columnTasks.length}
              </Badge>
            </div>

            <div className="space-y-3">
              {columnTasks.map((task) => (
                <div
                  key={task.id}
                  onClick={() => onTaskClick(task)}
                  className="bg-white p-4 rounded-xl border border-stone-200 shadow-sm hover:shadow-md hover:border-stone-300 transition-all cursor-pointer group"
                >
                  <div className="flex items-start justify-between gap-2 mb-2">
                    <h4 className="font-medium text-stone-900 leading-tight">{task.title}</h4>
                  </div>
                  
                  {task.description && (
                    <p className="text-sm text-stone-500 line-clamp-2 mb-4">
                      {task.description}
                    </p>
                  )}

                  <div className="flex items-center justify-between mt-4">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className={cn("capitalize text-[10px] px-1.5 py-0 h-5", getPriorityColor(task.priority))}>
                        {task.priority}
                      </Badge>
                      {task.due_date && (
                        <div className="flex items-center text-xs text-stone-400">
                          <Calendar className="w-3 h-3 mr-1" />
                          {format(new Date(task.due_date), 'MMM d')}
                        </div>
                      )}
                    </div>
                    
                    {/* Quick status actions */}
                    <div className="opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-1">
                      {column.id !== 'todo' && (
                        <button 
                          onClick={(e) => handleStatusChange(e, task.id, 'todo')}
                          className="p-1 text-stone-400 hover:text-stone-600 rounded hover:bg-stone-100"
                          title="Move to To Do"
                        >
                          <Circle className="w-4 h-4" />
                        </button>
                      )}
                      {column.id !== 'in_progress' && (
                        <button 
                          onClick={(e) => handleStatusChange(e, task.id, 'in_progress')}
                          className="p-1 text-stone-400 hover:text-orange-500 rounded hover:bg-orange-50"
                          title="Move to In Progress"
                        >
                          <ArrowRightCircle className="w-4 h-4" />
                        </button>
                      )}
                      {column.id !== 'done' && (
                        <button 
                          onClick={(e) => handleStatusChange(e, task.id, 'done')}
                          className="p-1 text-stone-400 hover:text-emerald-500 rounded hover:bg-emerald-50"
                          title="Move to Done"
                        >
                          <CheckCircle2 className="w-4 h-4" />
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              ))}
              
              {columnTasks.length === 0 && (
                <div className="text-center py-8 border-2 border-dashed border-stone-200 rounded-xl">
                  <p className="text-sm text-stone-400">No tasks</p>
                </div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
