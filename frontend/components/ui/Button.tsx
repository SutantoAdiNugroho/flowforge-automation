import { ButtonHTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/utils";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "danger" | "ghost" | "success";
  size?: "sm" | "md" | "lg";
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "primary", size = "md", ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(
          "inline-flex items-center justify-center rounded-lg font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 disabled:pointer-events-none disabled:opacity-50",
          variant === "primary" && "bg-blue-600 text-white hover:bg-blue-700 shadow-sm",
          variant === "secondary" && "bg-white text-gray-700 border border-gray-300 hover:bg-gray-50 shadow-sm",
          variant === "danger" && "bg-red-600 text-white hover:bg-red-700 shadow-sm",
          variant === "success" && "bg-green-600 text-white hover:bg-green-700 shadow-sm",
          variant === "ghost" && "hover:bg-gray-100 text-gray-700",
          size === "sm" && "h-8 px-3 text-xs",
          size === "md" && "h-10 px-4 py-2 text-sm",
          size === "lg" && "h-12 px-8 text-base",
          className
        )}
        {...props}
      />
    );
  }
);
Button.displayName = "Button";
