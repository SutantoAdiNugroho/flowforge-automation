"use client";

import { useEffect, useState, useRef } from "react";
import { fetchApi } from "@/lib/api";
import Link from "next/link";
import { useAuth } from "@/contexts/AuthContext";
import { FiEdit2, FiTrash2, FiCheck, FiX, FiSave, FiXCircle, FiPlus } from "react-icons/fi";

import { Button } from "@/components/ui/Button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/Table";
import { Badge } from "@/components/ui/Badge";
import { Card, CardContent } from "@/components/ui/Card";
import { Select } from "@/components/ui/Select";

export default function UsersPage() {
  const [users, setUsers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { user } = useAuth();
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingRole, setEditingRole] = useState("");
  const [editingIsActive, setEditingIsActive] = useState(true);
  const hasInitialized = useRef(false);

  const fetchUsers = async () => {
    try {
      const data = await fetchApi("/users?page=1&page_size=50");
      setUsers(data.content || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (hasInitialized.current) return;
    hasInitialized.current = true;
    fetchUsers();

    const handleFocus = () => {
      fetchUsers();
    };

    window.addEventListener("focus", handleFocus);
    return () => window.removeEventListener("focus", handleFocus);
  }, []);

  const handleEdit = (userId: string, role: string, isActive: boolean) => {
    setEditingId(userId);
    setEditingRole(role);
    setEditingIsActive(isActive);
  };

  const handleSave = async (userId: string) => {
    try {
      await fetchApi(`/users/${userId}`, {
        method: "PUT",
        body: JSON.stringify({
          role: editingRole,
          is_active: editingIsActive,
        }),
      });
      setEditingId(null);
      fetchUsers();
    } catch (err) {
      alert("Failed to update user");
    }
  };

  const handleCancel = () => {
    setEditingId(null);
  };

  const handleDelete = async (userId: string) => {
    if (!confirm("Are you sure you want to delete this user?")) return;
    try {
      await fetchApi(`/users/${userId}`, { method: "DELETE" });
      fetchUsers();
    } catch (err) {
      alert("Failed to delete user");
    }
  };

  if (loading) return <div className="text-gray-500">Loading...</div>;

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900 tracking-tight">Users Management</h1>
        <Link href="/dashboard/users/new">
          <Button className="flex items-center gap-2">
            <FiPlus className="w-4 h-4" />
            Create User
          </Button>
        </Link>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Email</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {users.map((usr) => (
                <TableRow key={usr.id}>
                  <TableCell className="font-medium text-gray-900">{usr.email}</TableCell>
                  <TableCell>
                    {editingId === usr.id ? (
                      <Select value={editingRole} onChange={(e) => setEditingRole(e.target.value)} className="h-8">
                        <option value="admin">Admin</option>
                        <option value="editor">Editor</option>
                        <option value="viewer">Viewer</option>
                      </Select>
                    ) : (
                      <span className="text-sm text-gray-600 capitalize">{usr.role}</span>
                    )}
                  </TableCell>
                  <TableCell>
                    {editingId === usr.id ? (
                      <label className="flex items-center gap-2">
                        <input
                          type="checkbox"
                          checked={editingIsActive}
                          onChange={(e) => setEditingIsActive(e.target.checked)}
                          className="w-4 h-4"
                        />
                        <span className="text-sm">{editingIsActive ? "Active" : "Inactive"}</span>
                      </label>
                    ) : (
                      <Badge variant={usr.is_active ? "success" : "danger"} className="flex items-center gap-1 w-fit">
                        {usr.is_active ? (
                          <>
                            <FiCheck className="w-3 h-3" />
                            Active
                          </>
                        ) : (
                          <>
                            <FiX className="w-3 h-3" />
                            Inactive
                          </>
                        )}
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell className="text-right space-x-2">
                    {editingId === usr.id ? (
                      <>
                        <Button
                          variant="primary"
                          size="sm"
                          onClick={() => handleSave(usr.id)}
                          className="inline-flex items-center gap-1"
                        >
                          <FiSave className="w-4 h-4" />
                          Save
                        </Button>
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={handleCancel}
                          className="inline-flex items-center gap-1"
                        >
                          <FiXCircle className="w-4 h-4" />
                          Cancel
                        </Button>
                      </>
                    ) : (
                      <>
                        {usr.id !== user?.id && (
                          <>
                            <Button
                              variant="secondary"
                              size="sm"
                              onClick={() => handleEdit(usr.id, usr.role, usr.is_active)}
                              className="inline-flex items-center gap-1"
                            >
                              <FiEdit2 className="w-4 h-4" />
                              Edit
                            </Button>
                            <Button
                              variant="danger"
                              size="sm"
                              onClick={() => handleDelete(usr.id)}
                              className="inline-flex items-center gap-1"
                            >
                              <FiTrash2 className="w-4 h-4" />
                              Delete
                            </Button>
                          </>
                        )}
                      </>
                    )}
                  </TableCell>
                </TableRow>
              ))}
              {users.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="text-center text-gray-500 py-8">
                    No users found
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
