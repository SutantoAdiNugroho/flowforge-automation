"use client";

import { useEffect, useState, use } from "react";
import { fetchApi } from "@/lib/api";
import { useAuth } from "@/contexts/AuthContext";
import Link from "next/link";

import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { Table, TableBody, TableCell, TableHead, TableHeader as UITableHeader, TableRow } from "@/components/ui/Table";
import { useSSE } from "@/hooks/useSSE";

export default function WorkflowDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const resolvedParams = use(params);
  const id = resolvedParams.id;
  const [workflow, setWorkflow] = useState<any>(null);
  const [runs, setRuns] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { hasRole } = useAuth();

  const fetchData = async () => {
    try {
      const [wfData, runsData] = await Promise.all([
        fetchApi(`/workflows/${id}`),
        fetchApi(`/workflows/${id}/runs?page=1&page_size=10`)
      ]);
      setWorkflow(wfData);
      setRuns(runsData.content || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [id]);

  useSSE("/events", (event) => {
    if (event.type === "run_update") {
      setRuns((prev) => {
        const idx = prev.findIndex((r) => r.id === event.payload.run_id);
        if (idx !== -1) {
          const newRuns = [...prev];
          newRuns[idx] = { ...newRuns[idx], status: event.payload.status };
          return newRuns;
        } else {
          // A new run might have been created, let's just refetch
          fetchData();
          return prev;
        }
      });
    }
  });

  const handleTrigger = async () => {
    try {
      await fetchApi(`/workflows/${id}/runs`, { method: "POST", body: "{}" });
      alert("Run triggered");
      fetchData();
    } catch (err: any) {
      alert("Failed to trigger: " + err.message);
    }
  };

  if (loading) return <div className="text-gray-500">Loading...</div>;
  if (!workflow) return <div className="text-red-500">Workflow not found</div>;

  const canEdit = hasRole(["admin", "editor"]);

  return (
    <div className="space-y-6">
      <Card>
        <CardContent className="p-6">
          <div className="flex justify-between items-start">
            <div>
              <h1 className="text-2xl font-bold text-gray-900 tracking-tight">{workflow.name}</h1>
              <p className="text-gray-500 mt-1">{workflow.description}</p>
              <div className="mt-4 flex space-x-3 text-sm text-gray-600">
                <Badge variant="default">Type: {workflow.trigger_type}</Badge>
                <Badge variant="default">Version: {workflow.version}</Badge>
                <Badge variant={workflow.is_active ? "success" : "danger"}>
                  {workflow.is_active ? "Active" : "Inactive"}
                </Badge>
              </div>
            </div>
            <div className="flex space-x-2">
              {canEdit && (
                <Button variant="primary" className="bg-green-600 hover:bg-green-700" onClick={handleTrigger}>
                  Trigger Run
                </Button>
              )}
              <Link href="/dashboard">
                <Button variant="secondary">Back</Button>
              </Link>
            </div>
          </div>

          <div className="mt-8">
            <h2 className="text-lg font-semibold mb-2 text-gray-900">Definition (JSON)</h2>
            <pre className="bg-gray-50 p-4 rounded-md text-sm overflow-x-auto text-gray-800 border border-gray-200">
              {JSON.stringify(workflow.definition, null, 2)}
            </pre>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Recent Runs</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <UITableHeader>
              <TableRow>
                <TableHead>Run ID</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Started</TableHead>
              </TableRow>
            </UITableHeader>
            <TableBody>
              {runs.map((r) => (
                <TableRow key={r.id}>
                  <TableCell className="font-mono text-sm text-blue-600 hover:underline">
                    <Link href={`/dashboard/runs/${r.id}`}>
                      {r.id.split('-')[0]}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <Badge variant={
                      r.status === 'success' ? 'success' : 
                      r.status === 'failed' ? 'danger' : 
                      'default'
                    }>
                      {r.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-gray-500">
                    {new Date(r.created_at).toLocaleString()}
                  </TableCell>
                </TableRow>
              ))}
              {runs.length === 0 && (
                <TableRow>
                  <TableCell colSpan={3} className="text-center text-gray-500 py-8">
                    No runs yet
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
