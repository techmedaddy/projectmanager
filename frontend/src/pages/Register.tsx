import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Link, useNavigate } from 'react-router-dom';
import { fetchApi, ApiError } from '../lib/api';
import { Button } from '../components/ui/button.tsx';
import { Input } from '../components/ui/input.tsx';
import { Label } from '../components/ui/label.tsx';
import { Card, CardContent, CardFooter } from '../components/ui/card.tsx';
import { CheckCircle2, Loader2 } from 'lucide-react';
import { toast } from 'sonner';

const registerSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
  email: z.string().email('Please enter a valid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
});

type RegisterFormValues = z.infer<typeof registerSchema>;

export function Register() {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (data: RegisterFormValues) => {
    setIsLoading(true);
    try {
      await fetchApi('/auth/register', {
        method: 'POST',
        body: JSON.stringify(data),
      });

      toast.success('Account created successfully. Please sign in.');
      navigate('/login');
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.fields) {
          Object.entries(error.fields).forEach(([field, message]) => {
            setError(field as keyof RegisterFormValues, { message });
          });
        } else {
          toast.error(error.message);
        }
      } else {
        toast.error('An unexpected error occurred');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-4 bg-gradient-to-br from-slate-950 via-indigo-950 to-slate-900">
      <div className="w-full max-w-md space-y-8">
        <div className="flex flex-col items-center text-center">
          <div className="w-14 h-14 rounded-2xl border border-cyan-300/40 bg-cyan-500/10 shadow-[0_0_26px_rgba(34,211,238,0.25)] flex items-center justify-center mb-6">
            <CheckCircle2 className="w-7 h-7 text-cyan-300" />
          </div>
          <h1 className="text-3xl font-semibold tracking-tight text-white">Create an account</h1>
          <p className="text-slate-300 mt-2">Get started with projectmanager today.</p>
        </div>

        <Card className="border-white/10 bg-slate-900/65 backdrop-blur-sm shadow-2xl shadow-black/30">
          <form onSubmit={handleSubmit(onSubmit)}>
            <CardContent className="pt-6 space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name" className="text-slate-200">Full Name</Label>
                <Input
                  id="name"
                  placeholder="Jane Doe"
                  {...register('name')}
                  className={errors.name ? 'border-rose-500 focus-visible:ring-rose-500' : 'bg-slate-800/80 border-white/15 text-slate-100'}
                />
                {errors.name && <p className="text-sm text-rose-400">{errors.name.message}</p>}
              </div>
              <div className="space-y-2">
                <Label htmlFor="email" className="text-slate-200">Email</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="name@example.com"
                  {...register('email')}
                  className={errors.email ? 'border-rose-500 focus-visible:ring-rose-500' : 'bg-slate-800/80 border-white/15 text-slate-100'}
                />
                {errors.email && <p className="text-sm text-rose-400">{errors.email.message}</p>}
              </div>
              <div className="space-y-2">
                <Label htmlFor="password" className="text-slate-200">Password</Label>
                <Input
                  id="password"
                  type="password"
                  {...register('password')}
                  className={errors.password ? 'border-rose-500 focus-visible:ring-rose-500' : 'bg-slate-800/80 border-white/15 text-slate-100'}
                />
                {errors.password && <p className="text-sm text-rose-400">{errors.password.message}</p>}
              </div>
            </CardContent>
            <CardFooter className="flex flex-col space-y-4">
              <Button type="submit" className="w-full bg-cyan-400 hover:bg-cyan-300 text-slate-950" disabled={isLoading}>
                {isLoading ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : null}
                Create account
              </Button>
              <div className="text-sm text-center text-slate-300">
                Already have an account?{' '}
                <Link to="/login" className="text-cyan-300 font-medium hover:text-cyan-200 hover:underline">
                  Sign in
                </Link>
              </div>
            </CardFooter>
          </form>
        </Card>
      </div>
    </div>
  );
}
