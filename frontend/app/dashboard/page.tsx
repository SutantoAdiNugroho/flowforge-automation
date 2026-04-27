"use client";

import { useEffect, useState, useRef } from "react";
import { fetchApi } from "@/lib/api";
import Link from "next/link";
import { useAuth } from "@/contexts/AuthContext";
import { FiPlus, FiEye, FiTrash2, FiCheck, FiX } from "react-icons/fi";

import { Button } from "@/components/ui/Button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/Table";
import { Badge } from "@/components/ui/Badge";
import { Card, CardContent } from "@/components/ui/Card";

export default function WorkflowsPage() {
  const [workflows, setWorkflows] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { hasRole } = useAuth();
  const hasInitialized = useRef(false);

  const fetchWorkflows = async () => {
    try {
      const data = await fetchApi("/workflows?page=1&page_size=50");
      setWorkflows(data.content || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (hasInitialized.current) return;
    hasInitialized.current = true;
    fetchWorkflows();

    const handleFocus = () => {
      fetchWorkflows();
    };

    window.addEventListener("focus", handleFocus);
    return () => window.removeEventListener("focus", handleFocus);
  }, []);

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure?")) return;
    try {
      await fetchApi(`/workflows/${id}`, { method: "DELETE" });
      fetchWorkflows();
    } catch (err) {
      alert("Failed to delete");
    }
  };

  if (loading) return <div className="text-gray-500">Loading...</div>;

  const canEdit = hasRole(["admin", "editor"]);
  const canDelete = hasRole(["admin"]);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900 tracking-tight">Workflows</h1>
        {canEdit && (
          <Link href="/dashboard/workflows/new">
            <Button className="flex items-center gap-2">
              <FiPlus className="w-4 h-4" />
              Create Workflow
            </Button>
          </Link>
        )}
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Trigger</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {workflows.map((wf) => (
                <TableRow key={wf.id}>
                  <TableCell className="font-medium text-gray-900">{wf.name}</TableCell>
                  <TableCell className="text-gray-500">{wf.trigger_type}</TableCell>
                  <TableCell>
                    <Badge variant={wf.is_active ? "success" : "danger"} className="flex items-center gap-1 w-fit">
                      {wf.is_active ? (
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
                  </TableCell>
                  <TableCell className="text-right space-x-2">
                    <Link href={`/dashboard/workflows/${wf.id}`}>
                      <Button variant="secondary" size="sm" className="inline-flex items-center gap-1">
                        <FiEye className="w-4 h-4" />
                        View
                      </Button>
                    </Link>
                    {canDelete && (
                      <Button variant="danger" size="sm" onClick={() => handleDelete(wf.id)} className="inline-flex items-center gap-1">
                        <FiTrash2 className="w-4 h-4" />
                        Delete
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
              {workflows.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="text-center text-gray-500 py-8">
                    No workflows found
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
