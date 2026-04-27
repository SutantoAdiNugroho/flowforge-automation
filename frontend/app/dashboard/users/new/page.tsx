"use client";

import { useState } from "react";
import { fetchApi } from "@/lib/api";
import { useRouter } from "next/navigation";
import { useAuth } from "@/contexts/AuthContext";
import { FiUserPlus, FiMail, FiLock, FiUserCheck } from "react-icons/fi";

import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Select } from "@/components/ui/Select";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";

export default function CreateUserPage() {
  const router = useRouter();
  const { hasRole } = useAuth();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("viewer");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  if (!hasRole(["admin"])) {
    return <div className="p-4 text-red-600">Access Denied</div>;
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await fetchApi("/users", {
        method: "POST",
        body: JSON.stringify({
          email,
          password,
          role,
        }),
      });
      router.refresh();
      router.push("/dashboard/users");
    } catch (err: any) {
      setError(err.message || "Failed to create user");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <FiUserPlus className="w-6 h-6 text-blue-600" />
            <CardTitle>Create User</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          {error && <div className="mb-4 text-sm text-red-600 bg-red-50 p-3 rounded-md border border-red-200">{error}</div>}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-1">
                <FiMail className="w-4 h-4" />
                Email
              </label>
              <Input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>

            <div>
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-1">
                <FiLock className="w-4 h-4" />
                Password (minimum 6 characters)
              </label>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                minLength={6}
                required
              />
            </div>

            <div>
              <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-1">
                <FiUserCheck className="w-4 h-4" />
                Role
              </label>
              <Select value={role} onChange={(e) => setRole(e.target.value)}>
                <option value="admin">Admin</option>
                <option value="editor">Editor</option>
                <option value="viewer">Viewer</option>
              </Select>
            </div>

            <div className="flex justify-end space-x-2 pt-4">
              <Button type="button" variant="secondary" onClick={() => router.back()}>
                Cancel
              </Button>
              <Button type="submit" disabled={loading} className="flex items-center gap-2">
                <FiUserPlus className="w-4 h-4" />
                {loading ? "Creating..." : "Create User"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
