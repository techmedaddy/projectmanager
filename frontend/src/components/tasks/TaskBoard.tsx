import React from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Task, TaskStatus } from '../../types';
import { fetchApi } from '../../lib/api';
import { toast } from 'sonner';
import { format } from 'date-fns';
import { Calendar, CheckCircle2, Circle, ArrowRightCircle } from 'lucide-react';
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

  const matchesTaskFilters = (task: Task, statusFilter: unknown, assigneeFilter: unknown) => {
    const statusMatches = statusFilter === 'all' || task.status === statusFilter;
    const assigneeMatches = assigneeFilter === 'all' || task.assignee_id === assigneeFilter;
    return statusMatches && assigneeMatches;
  };

  const updateStatusMutation = useMutation({
    mutationFn: ({ taskId, status }: { taskId: string; status: TaskStatus }) =>
      fetchApi(`/tasks/${taskId}`, {
        method: 'PATCH',
        body: JSON.stringify({ status }),
      }),
    onMutate: async ({ taskId, status }) => {
      await queryClient.cancelQueries({ queryKey: ['project-tasks', projectId] });
      const previousTaskQueries = queryClient.getQueriesData({ queryKey: ['project-tasks', projectId] }) as Array<[
        readonly unknown[],
        any,
      ]>;
      const previousProjectDetail = queryClient.getQueryData(['project-detail', projectId]);

      previousTaskQueries.forEach(([queryKey, cachedData]) => {
        const keyParts = Array.isArray(queryKey) ? queryKey : [];
        const statusFilter = keyParts[2] ?? 'all';
        const assigneeFilter = keyParts[3] ?? 'all';

        queryClient.setQueryData(queryKey, (old: any) => {
          const sourceTasks: Task[] = old?.items ?? old?.tasks ?? (cachedData as any)?.items ?? (cachedData as any)?.tasks;
          if (!Array.isArray(sourceTasks)) return old;

          const updatedTasks = sourceTasks
            .map((task) => (task.id === taskId ? { ...task, status } : task))
            .filter((task) => matchesTaskFilters(task, statusFilter, assigneeFilter));

          return {
            ...(old ?? cachedData),
            items: updatedTasks,
          };
        });
      });

      queryClient.setQueryData(['project-detail', projectId], (old: any) => {
        if (!old?.tasks) return old;
        return {
          ...old,
          tasks: old.tasks.map((t: Task) => (t.id === taskId ? { ...t, status } : t)),
        };
      });

      return { previousTaskQueries, previousProjectDetail };
    },
    onError: (_err, _variables, context) => {
      context?.previousTaskQueries?.forEach(([queryKey, data]: [unknown, unknown]) => {
        queryClient.setQueryData(queryKey as readonly unknown[], data);
      });
      queryClient.setQueryData(['project-detail', projectId], context?.previousProjectDetail);
      toast.error('Failed to update task status');
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['project-tasks', projectId] });
      queryClient.invalidateQueries({ queryKey: ['project-detail', projectId] });
    },
  });

  const handleStatusChange = (e: React.MouseEvent, taskId: string, newStatus: TaskStatus) => {
    e.stopPropagation();
    updateStatusMutation.mutate({ taskId, status: newStatus });
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'bg-rose-500/20 text-rose-200 border-rose-300/30';
      case 'medium': return 'bg-amber-500/20 text-amber-200 border-amber-300/30';
      case 'low': return 'bg-emerald-500/20 text-emerald-200 border-emerald-300/30';
      default: return 'bg-slate-500/20 text-slate-200 border-slate-300/30';
    }
  };

  return (
    <div className="flex gap-6 h-full items-start">
      {COLUMNS.map((column) => {
        const columnTasks = tasks.filter((t) => t.status === column.id);
        const Icon = column.icon;

        return (
          <div key={column.id} className="flex-1 min-w-[300px] rounded-2xl p-4 border border-white/10 bg-slate-900/45 backdrop-blur-sm">
            <div className="flex items-center justify-between mb-4 px-1">
              <div className="flex items-center gap-2 text-slate-100 font-medium">
                <Icon className={cn("w-4 h-4", 
                  column.id === 'todo' ? 'text-slate-400' : 
                  column.id === 'in_progress' ? 'text-cyan-300' : 
                  'text-emerald-300'
                )} />
                {column.title}
              </div>
              <Badge variant="secondary" className="bg-slate-700/60 text-slate-200 hover:bg-slate-700/60 border border-white/10">
                {columnTasks.length}
              </Badge>
            </div>

            <div className="space-y-3">
              {columnTasks.map((task) => (
                <div
                  key={task.id}
                  onClick={() => onTaskClick(task)}
                  className="bg-slate-800/70 p-4 rounded-xl border border-white/10 shadow-sm hover:shadow-lg hover:border-cyan-300/30 transition-all cursor-pointer group"
                >
                  <div className="flex items-start justify-between gap-2 mb-2">
                    <h4 className="font-medium text-slate-100 leading-tight">{task.title}</h4>
                  </div>
                  
                  {task.description && (
                    <p className="text-sm text-slate-300 line-clamp-2 mb-4">
                      {task.description}
                    </p>
                  )}

                  <div className="flex items-center justify-between mt-4">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className={cn("capitalize text-[10px] px-1.5 py-0 h-5", getPriorityColor(task.priority))}>
                        {task.priority}
                      </Badge>
                      {task.due_date && (
                        <div className="flex items-center text-xs text-slate-400">
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
                          className="p-1 text-slate-400 hover:text-slate-200 rounded hover:bg-white/10"
                          title="Move to To Do"
                        >
                          <Circle className="w-4 h-4" />
                        </button>
                      )}
                      {column.id !== 'in_progress' && (
                        <button 
                          onClick={(e) => handleStatusChange(e, task.id, 'in_progress')}
                          className="p-1 text-slate-400 hover:text-cyan-300 rounded hover:bg-cyan-500/10"
                          title="Move to In Progress"
                        >
                          <ArrowRightCircle className="w-4 h-4" />
                        </button>
                      )}
                      {column.id !== 'done' && (
                        <button 
                          onClick={(e) => handleStatusChange(e, task.id, 'done')}
                          className="p-1 text-slate-400 hover:text-emerald-300 rounded hover:bg-emerald-500/10"
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
                <div className="text-center py-8 border-2 border-dashed border-white/15 rounded-xl">
                  <p className="text-sm text-slate-400">No tasks</p>
                </div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
