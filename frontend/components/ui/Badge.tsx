import { HTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/utils";

interface BadgeProps extends HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "success" | "danger" | "warning";
}

export const Badge = forwardRef<HTMLDivElement, BadgeProps>(
  ({ className, variant = "default", ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-gray-950 focus:ring-offset-2",
          variant === "default" && "border-transparent bg-gray-100 text-gray-900 hover:bg-gray-200",
          variant === "success" && "border-transparent bg-green-100 text-green-800 hover:bg-green-200",
          variant === "danger" && "border-transparent bg-red-100 text-red-800 hover:bg-red-200",
          variant === "warning" && "border-transparent bg-yellow-100 text-yellow-800 hover:bg-yellow-200",
          className
        )}
        {...props}
      />
    );
  }
);
Badge.displayName = "Badge";
