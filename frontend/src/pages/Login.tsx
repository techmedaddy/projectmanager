import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { fetchApi, ApiError } from '../lib/api';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Card, CardContent, CardFooter } from '../components/ui/card';
import { CheckCircle2, Loader2 } from 'lucide-react';
import { toast } from 'sonner';

const loginSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
  password: z.string().min(1, 'Password is required'),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export function Login() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [isLoading, setIsLoading] = useState(false);

  const from = location.state?.from?.pathname || '/';

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: 'test@example.com',
      password: 'password123',
    },
  });

  const onSubmit = async (data: LoginFormValues) => {
    setIsLoading(true);
    try {
      const response = await fetchApi<{ access_token: string }>('/auth/login', {
        method: 'POST',
        body: JSON.stringify(data),
      });

      const meData = await fetchApi<{ user: { id: string; name: string; email: string } }>('/auth/me', {
        headers: {
          Authorization: `Bearer ${response.access_token}`,
        },
      });

      login(response.access_token, meData.user);
      toast.success('Welcome back!');
      navigate(from, { replace: true });
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.fields) {
          Object.entries(error.fields).forEach(([field, message]) => {
            setError(field as keyof LoginFormValues, { message });
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
    <div className="min-h-screen flex flex-col items-center justify-center p-4 bg-slate-100">
      <div className="w-full max-w-md space-y-8">
        <div className="flex flex-col items-center text-center">
          <div className="w-14 h-14 rounded-2xl border border-indigo-200 bg-indigo-50 flex items-center justify-center mb-6">
            <CheckCircle2 className="w-7 h-7 text-indigo-600" />
          </div>
          <h1 className="text-3xl font-semibold tracking-tight text-slate-900">Welcome back</h1>
          <p className="text-slate-600 mt-2">Enter your credentials to access your tasks.</p>
        </div>

        <Card className="border-slate-200 bg-white shadow-sm">
          <form onSubmit={handleSubmit(onSubmit)}>
            <CardContent className="pt-6 space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email" className="text-slate-700">Email</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="name@example.com"
                  {...register('email')}
                  className={errors.email ? 'border-rose-500 focus-visible:ring-rose-500' : ''}
                />
                {errors.email && <p className="text-sm text-rose-500">{errors.email.message}</p>}
              </div>
              <div className="space-y-2">
                <Label htmlFor="password" className="text-slate-700">Password</Label>
                <Input
                  id="password"
                  type="password"
                  {...register('password')}
                  className={errors.password ? 'border-rose-500 focus-visible:ring-rose-500' : ''}
                />
                {errors.password && <p className="text-sm text-rose-500">{errors.password.message}</p>}
              </div>
            </CardContent>
            <CardFooter className="flex flex-col space-y-4">
              <Button type="submit" className="w-full bg-slate-900 hover:bg-slate-800 text-white" disabled={isLoading}>
                {isLoading ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : null}
                Sign in
              </Button>
              <div className="text-sm text-center text-slate-600">
                Don't have an account?{' '}
                <Link to="/register" className="text-indigo-700 font-medium hover:text-indigo-800 hover:underline">
                  Create one
                </Link>
              </div>
            </CardFooter>
          </form>
        </Card>
      </div>
    </div>
  );
}
