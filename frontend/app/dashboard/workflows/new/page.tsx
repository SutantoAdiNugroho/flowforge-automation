"use client";

import { useState } from "react";
import { fetchApi } from "@/lib/api";
import { useRouter } from "next/navigation";
import { useAuth } from "@/contexts/AuthContext";

import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Select } from "@/components/ui/Select";
import { Textarea } from "@/components/ui/Textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";

export default function CreateWorkflowPage() {
  const router = useRouter();
  const { hasRole } = useAuth();
  
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [triggerType, setTriggerType] = useState("manual");
  const [cronExpression, setCronExpression] = useState("");
  const [definition, setDefinition] = useState("{\n  \"steps\": []\n}");
  const [error, setError] = useState("");

  if (!hasRole(["admin", "editor"])) {
    return <div className="p-4 text-red-600">Access Denied</div>;
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    let parsedDef;
    try {
      parsedDef = JSON.parse(definition);
    } catch (err) {
      setError("Invalid JSON definition");
      return;
    }

    try {
      await fetchApi("/workflows", {
        method: "POST",
        body: JSON.stringify({
          name,
          description,
          trigger_type: triggerType,
          cron_expression: triggerType === "cron" ? cronExpression : "",
          definition: parsedDef,
          is_active: true,
        }),
      });
      router.refresh();
      router.push("/dashboard");
    } catch (err: any) {
      setError(err.message || "Failed to create");
    }
  };

  return (
    <div className="max-w-3xl mx-auto">
      <Card>
        <CardHeader>
          <CardTitle>Create Workflow</CardTitle>
        </CardHeader>
        <CardContent>
          {error && <div className="mb-4 text-sm text-red-600 bg-red-50 p-3 rounded-md border border-red-200">{error}</div>}
          
          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label="Name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
            
            <Textarea
              label="Description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
            
            <Select
              label="Trigger Type"
              value={triggerType}
              onChange={(e) => setTriggerType(e.target.value)}
            >
              <option value="manual">Manual</option>
              <option value="cron">Cron</option>
              <option value="webhook">Webhook</option>
            </Select>

            {triggerType === "cron" && (
              <Input
                label="Cron Expression"
                type="text"
                placeholder="e.g. 0 9 * * *"
                value={cronExpression}
                onChange={(e) => setCronExpression(e.target.value)}
                required
              />
            )}
            
            <Textarea
              label="DAG Definition (JSON)"
              className="font-mono text-sm h-64"
              value={definition}
              onChange={(e) => setDefinition(e.target.value)}
              required
            />
            
            <div className="flex justify-end space-x-2 pt-4">
              <Button type="button" variant="secondary" onClick={() => router.back()}>Cancel</Button>
              <Button type="submit">Create</Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
