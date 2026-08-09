import { fireEvent, render, screen } from "@testing-library/react";
import { CircleAlert } from "lucide-react";
import { describe, expect, it, vi } from "vitest";

import { EmptyState } from "./EmptyState";

describe("EmptyState", () => {
  it("renders its message without an action", () => {
    render(<EmptyState icon={CircleAlert} title="No memories" description="Create the first memory." />);

    expect(screen.getByRole("heading", { name: "No memories" })).toBeInTheDocument();
    expect(screen.getByText("Create the first memory.")).toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("invokes the optional action", () => {
    const onClick = vi.fn();
    render(
      <EmptyState
        icon={CircleAlert}
        title="No memories"
        description="Create the first memory."
        action={{ label: "Create", onClick }}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Create" }));
    expect(onClick).toHaveBeenCalledOnce();
  });
});
