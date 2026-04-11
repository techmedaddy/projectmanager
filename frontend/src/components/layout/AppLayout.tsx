import { Outlet, Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { LogOut, CheckCircle2 } from 'lucide-react';
import { Button } from '../ui/button';

export function AppLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen flex flex-col text-slate-900 font-sans">
      <header className="sticky top-0 z-20 border-b border-slate-200/80 bg-white/90 backdrop-blur-lg">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2.5 text-slate-900 hover:opacity-90 transition-opacity">
            <span className="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-indigo-200 bg-indigo-50">
              <CheckCircle2 className="w-5 h-5 text-indigo-600" />
            </span>
            <span className="font-semibold text-lg tracking-tight">projectmanager</span>
          </Link>

          <div className="flex items-center gap-2 sm:gap-4">
            <span className="text-sm font-medium text-slate-600 hidden sm:inline-block rounded-full border border-slate-200 bg-slate-50 px-3 py-1">
              {user?.name}
            </span>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleLogout}
              className="text-slate-600 hover:text-slate-900 hover:bg-slate-100"
            >
              <LogOut className="w-4 h-4 mr-2" />
              Logout
            </Button>
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>
    </div>
  );
}
