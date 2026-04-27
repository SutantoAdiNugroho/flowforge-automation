"use client";

import { useEffect, useState, use } from "react";
import { fetchApi } from "@/lib/api";
import Link from "next/link";
import { FiCheck, FiX, FiClock, FiLoader } from "react-icons/fi";
import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { Table, TableBody, TableCell, TableHead, TableHeader as UITableHeader, TableRow } from "@/components/ui/Table";
import { useSSE } from "@/hooks/useSSE";

export default function RunDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const resolvedParams = use(params);
  const id = resolvedParams.id;

  const [runData, setRunData] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  const fetchRun = async () => {
    try {
      const data = await fetchApi(`/runs/${id}`);
      setRunData(data);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRun();
  }, [id]);

  useSSE(`/events?run_id=${id}`, (event) => {
    if (event.type === "step_update") {
      setRunData((prev: any) => {
        if (!prev) return prev;

        const stepExists = prev.steps.some((s: any) => s.step_id === event.payload.step_id);

        if (stepExists) {
          const newSteps = prev.steps.map((step: any) =>
            step.step_id === event.payload.step_id
              ? {
                ...step,
                status: event.payload.status,
                error: event.payload.error,
                duration_ms: event.payload.duration_ms || step.duration_ms,
                updated_at: event.payload.timestamp
              }
              : step
          );
          return { ...prev, steps: newSteps };
        } else {
          const newStep = {
            id: event.payload.step_id,
            step_id: event.payload.step_id,
            step_name: event.payload.step_name,
            status: event.payload.status,
            error: event.payload.error,
            duration_ms: event.payload.duration_ms,
            updated_at: event.payload.timestamp
          };
          return { ...prev, steps: [...prev.steps, newStep] };
        }
      });
    } else if (event.type === "run_update") {
      setRunData((prev: any) => {
        if (!prev) return prev;
        return { ...prev, run: { ...prev.run, status: event.payload.status } };
      });
    }
  });

  const handleCancel = async () => {
    if (!confirm("Cancel this run?")) return;
    try {
      await fetchApi(`/runs/${id}/cancel`, { method: "POST" });
    } catch (err: any) {
      alert("Failed to cancel: " + err.message);
    }
  };

  if (loading) return <div className="text-gray-500">Loading run details...</div>;
  if (!runData) return <div className="text-red-500">Run not found</div>;

  const { run, steps } = runData;
  const isPendingOrRunning = run.status === "pending" || run.status === "running";

  return (
    <div className="space-y-6">
      <Card>
        <CardContent className="p-6">
          <div className="flex justify-between items-start">
            <div>
              <h1 className="text-2xl font-bold text-gray-900 tracking-tight">Run: {run.id.split('-')[0]}</h1>
              <p className="text-gray-500 mt-1">Workflow ID: {run.workflow_id}</p>
              <div className="mt-4 flex space-x-3">
                <Badge variant={
                  run.status === 'success' ? 'success' :
                    run.status === 'failed' ? 'danger' :
                      run.status === 'cancelled' ? 'warning' : 'default'
                }>
                  Status: {run.status}
                </Badge>
                <Badge variant="default">Trigger: {run.triggered_by}</Badge>
              </div>
            </div>
            <div className="flex space-x-2">
              {isPendingOrRunning && (
                <Button variant="danger" onClick={handleCancel}>Cancel Run</Button>
              )}
              <Link href={`/dashboard/workflows/${run.workflow_id}`}>
                <Button variant="secondary">Back to Workflow</Button>
              </Link>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Step Executions</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <UITableHeader>
              <TableRow>
                <TableHead>Step ID</TableHead>
                <TableHead>Step Name</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Duration</TableHead>
                <TableHead>Error</TableHead>
              </TableRow>
            </UITableHeader>
            <TableBody>
              {steps.map((s: any) => (
                <TableRow key={s.id}>
                  <TableCell className="font-mono text-sm text-gray-900">{s.step_id}</TableCell>
                  <TableCell className="text-gray-700">{s.step_name}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {s.status === 'success' && <FiCheck className="w-4 h-4 text-green-600" />}
                      {s.status === 'failed' && <FiX className="w-4 h-4 text-red-600" />}
                      {s.status === 'running' && <FiLoader className="w-4 h-4 text-blue-600 animate-spin" />}
                      {s.status === 'pending' && <FiClock className="w-4 h-4 text-gray-400" />}
                      {s.status === 'skipped' && <FiX className="w-4 h-4 text-gray-400" />}
                      <Badge variant={
                        s.status === 'success' ? 'success' :
                          s.status === 'failed' ? 'danger' :
                            s.status === 'running' ? 'warning' : 'default'
                      }>
                        {s.status}
                      </Badge>
                    </div>
                  </TableCell>
                  <TableCell className="text-gray-600 text-sm">
                    {s.duration_ms ? `${s.duration_ms}ms` : '-'}
                  </TableCell>
                  <TableCell className="text-red-600 text-sm">{s.error || "-"}</TableCell>
                </TableRow>
              ))}
              {steps.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-gray-500 py-8">
                    No steps executed yet.
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
