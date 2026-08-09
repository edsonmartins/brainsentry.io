import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { Pagination } from "./pagination";

describe("Pagination", () => {
  it("navigates using page buttons and a valid jump", () => {
    const onPageChange = vi.fn();
    render(<Pagination currentPage={4} totalPages={10} onPageChange={onPageChange} />);

    fireEvent.click(screen.getByRole("button", { name: "5" }));
    const jump = screen.getByRole("spinbutton");
    fireEvent.change(jump, { target: { value: "8" } });
    fireEvent.keyDown(jump, { key: "Enter" });

    expect(onPageChange).toHaveBeenNthCalledWith(1, 5);
    expect(onPageChange).toHaveBeenNthCalledWith(2, 8);
    expect(jump).toHaveValue(null);
  });

  it("changes page size and ignores an invalid jump", () => {
    const onPageChange = vi.fn();
    const onPageSizeChange = vi.fn();
    render(
      <Pagination
        currentPage={1}
        totalPages={2}
        totalItems={12}
        pageSize={10}
        onPageChange={onPageChange}
        showPageSizeSelector
        onPageSizeChange={onPageSizeChange}
      />,
    );

    fireEvent.change(screen.getByRole("combobox"), { target: { value: "25" } });
    const jump = screen.getByRole("spinbutton");
    fireEvent.change(jump, { target: { value: "3" } });
    fireEvent.keyDown(jump, { key: "Enter" });

    expect(onPageSizeChange).toHaveBeenCalledWith(25);
    expect(onPageChange).not.toHaveBeenCalled();
    expect(screen.getByText("Mostrando 1 a 10 de 12 itens")).toBeInTheDocument();
  });
});
