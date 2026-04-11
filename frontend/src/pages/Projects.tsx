import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { format } from 'date-fns';
import { ApiError, fetchApi } from '../lib/api';
import { Project, ProjectsResponse } from '../types';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Textarea } from '../components/ui/textarea';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from '../components/ui/dialog';
import { FolderKanban, Plus, Loader2, Calendar } from 'lucide-react';
import { toast } from 'sonner';

const projectSchema = z.object({
  name: z.string().min(1, 'Project name is required'),
  description: z.string().optional(),
});

type ProjectFormValues = z.infer<typeof projectSchema>;

export function Projects() {
  const queryClient = useQueryClient();
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [page, setPage] = useState(1);
  const limit = 6;

  const { data, isLoading, error, isFetching } = useQuery<ProjectsResponse>({
    queryKey: ['projects', page, limit],
    queryFn: () => fetchApi(`/projects?page=${page}&limit=${limit}`),
  });

  const createMutation = useMutation({
    mutationFn: (newProject: ProjectFormValues) =>
      fetchApi<Project>('/projects', {
        method: 'POST',
        body: JSON.stringify(newProject),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      setPage(1);
      setIsCreateOpen(false);
      reset();
      toast.success('Project created successfully');
    },
    onError: () => {
      toast.error('Failed to create project');
    },
  });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ProjectFormValues>({
    resolver: zodResolver(projectSchema),
  });

  const onSubmit = (formData: ProjectFormValues) => {
    createMutation.mutate(formData);
  };

  if (isLoading) {
    return (
      <div className="flex justify-center py-16">
        <Loader2 className="w-8 h-8 animate-spin text-cyan-300" />
      </div>
    );
  }

  if (error) {
    const apiError = error instanceof ApiError ? error : null;
    const status = apiError?.status;

    const message =
      status === 401
        ? 'Your session expired. Please sign in again.'
        : status === 403
          ? 'You do not have permission to view projects.'
          : status === 404
            ? 'Projects resource was not found.'
            : 'Failed to load projects. Please try again.';

    return (
      <div className="text-center py-12">
        <p className="text-rose-300">{message}</p>
        {status && <p className="text-xs text-slate-400 mt-1">Error code: {status}</p>}
      </div>
    );
  }

  const projects = data?.items ?? data?.projects ?? [];
  const meta = data?.meta;
  const totalPages = meta ? Math.max(1, Math.ceil(meta.total / meta.limit)) : 1;

  return (
    <div className="space-y-8">
      <section className="rounded-3xl border border-white/10 bg-gradient-to-r from-indigo-500/20 via-cyan-500/10 to-transparent p-6 shadow-[0_20px_50px_rgba(0,0,0,0.25)]">
        <div className="flex items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-semibold tracking-tight text-white">Projects</h1>
            <p className="text-slate-300 mt-1">Manage your workspaces and tasks.</p>
          </div>

          <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
            <DialogTrigger asChild>
              <Button className="bg-cyan-400 text-slate-950 hover:bg-cyan-300 shadow-lg shadow-cyan-500/30">
                <Plus className="w-4 h-4 mr-2" />
                New Project
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px] bg-slate-900 border-white/15 text-slate-100">
              <DialogHeader>
                <DialogTitle>Create Project</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 pt-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Name</Label>
                  <Input
                    id="name"
                    placeholder="e.g. Website Redesign"
                    {...register('name')}
                    className={errors.name ? 'border-rose-500' : ''}
                  />
                  {errors.name && <p className="text-sm text-rose-400">{errors.name.message}</p>}
                </div>
                <div className="space-y-2">
                  <Label htmlFor="description">Description (Optional)</Label>
                  <Textarea
                    id="description"
                    placeholder="Briefly describe the project..."
                    {...register('description')}
                    className="resize-none"
                  />
                </div>
                <DialogFooter>
                  <Button type="button" variant="outline" onClick={() => setIsCreateOpen(false)}>
                    Cancel
                  </Button>
                  <Button type="submit" disabled={createMutation.isPending} className="bg-cyan-400 text-slate-950 hover:bg-cyan-300">
                    {createMutation.isPending && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                    Create
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>
      </section>

      {projects.length === 0 ? (
        <div className="text-center py-24 border border-dashed border-white/20 rounded-3xl bg-slate-900/40 backdrop-blur">
          <FolderKanban className="w-12 h-12 text-cyan-300/70 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-white">No projects yet</h3>
          <p className="text-slate-300 mt-1 mb-6">Create your first project to start organizing tasks.</p>
          <Button onClick={() => setIsCreateOpen(true)} className="bg-cyan-400 text-slate-950 hover:bg-cyan-300">
            <Plus className="w-4 h-4 mr-2" />
            Create Project
          </Button>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {projects.map((project) => (
              <Link key={project.id} to={`/projects/${project.id}`} className="block group">
                <Card className="h-full rounded-2xl border-white/10 bg-slate-900/55 backdrop-blur-sm shadow-md transition-all duration-200 hover:-translate-y-1 hover:shadow-xl hover:border-cyan-300/40">
                  <CardHeader>
                    <CardTitle className="text-lg text-white group-hover:text-cyan-300 transition-colors">
                      {project.name}
                    </CardTitle>
                    <CardDescription className="line-clamp-2 mt-2 text-slate-300">
                      {project.description || 'No description provided.'}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-center text-xs text-slate-400 mt-4 border-t border-white/10 pt-3">
                      <Calendar className="w-3.5 h-3.5 mr-1.5" />
                      Created {format(new Date(project.created_at), 'MMM d, yyyy')}
                    </div>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>

          <div className="flex items-center justify-between rounded-2xl border border-white/10 bg-slate-900/45 p-3.5 shadow-sm backdrop-blur-sm">
            <p className="text-sm text-slate-300">
              Page <span className="font-medium text-white">{meta?.page ?? page}</span> of{' '}
              <span className="font-medium text-white">{totalPages}</span>
              {isFetching ? ' · Updating…' : ''}
            </p>
            <div className="flex items-center gap-2">
              <Button variant="outline" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={(meta?.page ?? page) <= 1 || isFetching}>
                Previous
              </Button>
              <Button variant="outline" onClick={() => setPage((p) => Math.min(totalPages, p + 1))} disabled={(meta?.page ?? page) >= totalPages || isFetching}>
                Next
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
