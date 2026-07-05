import type { ReactNode } from "react";

// Section standardizes the vertical rhythm and top divider shared by every
// content section, so spacing is consistent site-wide from one place.
export function Section({
  id,
  children,
  className = "",
}: {
  id?: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <section
      id={id}
      className={`scroll-mt-20 border-t border-border py-16 ${className}`}
    >
      {children}
    </section>
  );
}
