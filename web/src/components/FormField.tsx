import type { ReactNode } from "react";

interface FormFieldProps {
  label: string;
  helperText?: ReactNode;
  error?: string;
  children: ReactNode;
}

export default function FormField({ label, helperText, error, children }: FormFieldProps) {
  return (
    <div>
      <label className="block text-xs font-semibold uppercase tracking-wider text-text-muted mb-2">
        {label}
      </label>
      {children}
      {helperText && !error && (
        <p className="mt-2 text-sm text-text-muted">{helperText}</p>
      )}
      {error && (
        <p className="mt-2 text-sm text-danger">{error}</p>
      )}
    </div>
  );
}
