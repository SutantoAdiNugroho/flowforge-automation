"use client";

import { useEffect, useState } from "react";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:5000/api";

export function useSSE(endpoint: string, onMessage: (event: any) => void) {
  useEffect(() => {
    let controller = new AbortController();
    
    const connect = async () => {
      const token = localStorage.getItem("token") || "";
      try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
          headers: {
            Authorization: `Bearer ${token}`
          },
          signal: controller.signal
        });

        if (!response.body) return;

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { value, done } = await reader.read();
          if (done) break;
          
          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split("\n\n");
          buffer = lines.pop() || "";

          for (const line of lines) {
            if (line.startsWith("data: ")) {
              const dataStr = line.replace("data: ", "").trim();
              if (dataStr) {
                try {
                  const parsed = JSON.parse(dataStr);
                  onMessage(parsed);
                } catch (e) {
                  console.error("SSE parse error", e);
                }
              }
            }
          }
        }
      } catch (err: any) {
        if (err.name !== "AbortError") {
          console.error("SSE connection error", err);
          // Simple reconnect logic
          setTimeout(connect, 3000);
        }
      }
    };

    connect();

    return () => {
      controller.abort();
    };
  }, [endpoint]); // Using onMessage in dep array can cause infinite loops if not memoized, omitting it.
}
