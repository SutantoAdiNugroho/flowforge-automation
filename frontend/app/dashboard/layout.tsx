"use client";

import { useState, useEffect } from "react";
import { useAuth } from "@/contexts/AuthContext";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { FiMenu, FiX, FiLogOut, FiHome, FiUsers, FiGitlab, FiShield } from "react-icons/fi";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const { user, logout } = useAuth();
  const pathname = usePathname();
  const router = useRouter();
  const [sidebarOpen, setSidebarOpen] = useState(true);

  const isActive = (path: string) => pathname === path || pathname.startsWith(path + "/");

  useEffect(() => {
    if (user?.role === "super-admin" && pathname === "/dashboard") {
      router.push("/admin/tenants");
    }
  }, [user, pathname, router]);

  if (!user) return null;

  return (
    <div className="flex h-screen bg-gray-100">
      <aside className={`${sidebarOpen ? "w-64" : "w-20"} bg-gray-900 text-white flex flex-col transition-all duration-300`}>
        <div className="p-4 border-b border-gray-800 flex items-center justify-between">
          {sidebarOpen && (
            <div>
              <div className="flex items-center gap-2">
                <FiGitlab className="w-6 h-6 text-blue-400" />
                <h1 className="text-xl font-bold">FlowForge</h1>
              </div>
              <p className="text-xs text-gray-400 mt-2 truncate w-40">{user.email}</p>
              <p className="text-xs text-gray-500 uppercase mt-1">Role: {user.role}</p>
            </div>
          )}
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-2 hover:bg-gray-800 rounded transition-colors"
            aria-label="Toggle sidebar"
          >
            {sidebarOpen ? <FiX className="w-5 h-5" /> : <FiMenu className="w-5 h-5" />}
          </button>
        </div>

        <nav className="flex-1 p-4 space-y-2">
          {user.role !== "super-admin" && (
            <Link
              href="/dashboard"
              className={`flex items-center gap-3 p-2 rounded hover:bg-gray-800 transition-colors ${
                isActive("/dashboard") && pathname === "/dashboard" ? "bg-gray-800" : ""
              }`}
              title="Workflows"
            >
              <FiHome className="w-5 h-5 flex-shrink-0" />
              {sidebarOpen && <span>Workflows</span>}
            </Link>
          )}

          {user.role === "super-admin" && (
            <Link
              href="/admin/tenants"
              className={`flex items-center gap-3 p-2 rounded hover:bg-gray-800 transition-colors ${
                isActive("/admin") ? "bg-gray-800 text-red-400" : "text-red-400"
              }`}
              title="System Admin"
            >
              <FiShield className="w-5 h-5 flex-shrink-0" />
              {sidebarOpen && <span>System Admin</span>}
            </Link>
          )}

          {user.role === "admin" && (
            <Link
              href="/dashboard/users"
              className={`flex items-center gap-3 p-2 rounded hover:bg-gray-800 transition-colors ${
                isActive("/dashboard/users") ? "bg-gray-800" : ""
              }`}
              title="Users Management"
            >
              <FiUsers className="w-5 h-5 flex-shrink-0" />
              {sidebarOpen && <span>Users Management</span>}
            </Link>
          )}
        </nav>

        <div className="p-4 border-t border-gray-800">
          <button
            onClick={logout}
            className="w-full flex items-center gap-3 p-2 text-red-400 hover:bg-gray-800 rounded transition-colors justify-center"
            title="Logout"
          >
            <FiLogOut className="w-5 h-5 flex-shrink-0" />
            {sidebarOpen && <span>Logout</span>}
          </button>
        </div>
      </aside>

      <main className="flex-1 overflow-auto p-8">
        {children}
      </main>
    </div>
  );
}
