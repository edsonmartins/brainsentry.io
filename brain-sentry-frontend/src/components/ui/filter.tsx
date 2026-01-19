import * as React from "react";
import { Search, X, Filter, ChevronDown } from "lucide-react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";
import { Button } from "./button";

const inputVariants = cva(
  "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "",
        ghost: "border-transparent bg-transparent shadow-none",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement>,
    VariantProps<typeof inputVariants> {}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, variant, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(inputVariants({ variant, className }))}
        ref={ref}
        {...props}
      />
    );
  }
);
Input.displayName = "Input";

interface SearchInputProps extends Omit<InputProps, "value" | "onChange"> {
  value: string;
  onChange: (value: string) => void;
  onClear?: () => void;
  placeholder?: string;
  debounceMs?: number;
  className?: string;
}

function SearchInput({
  value,
  onChange,
  onClear,
  placeholder = "Buscar...",
  debounceMs = 300,
  className,
}: SearchInputProps) {
  const [localValue, setLocalValue] = React.useState(value);

  React.useEffect(() => {
    setLocalValue(value);
  }, [value]);

  React.useEffect(() => {
    const timer = setTimeout(() => {
      if (localValue !== value) {
        onChange(localValue);
      }
    }, debounceMs);

    return () => clearTimeout(timer);
  }, [localValue, debounceMs, onChange, value]);

  const handleClear = () => {
    setLocalValue("");
    onChange("");
    onClear?.();
  };

  return (
    <div className="relative">
      <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
      <Input
        type="text"
        value={localValue}
        onChange={(e) => setLocalValue(e.target.value)}
        placeholder={placeholder}
        className={cn("pl-9 pr-9", className)}
      />
      {localValue && (
        <button
          onClick={handleClear}
          className="absolute right-2.5 top-2.5 text-muted-foreground hover:text-foreground"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  );
}

interface FilterOption {
  value: string;
  label: string;
  count?: number;
}

interface FilterSelectProps {
  label: string;
  options: FilterOption[];
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

function FilterSelect({
  label,
  options,
  value,
  onChange,
  placeholder = "Todos",
  className,
}: FilterSelectProps) {
  const [isOpen, setIsOpen] = React.useState(false);
  const selectRef = React.useRef<HTMLDivElement>(null);

  const selectedOption = options.find((opt) => opt.value === value);

  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (selectRef.current && !selectRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <div ref={selectRef} className={cn("relative", className)}>
      <label className="text-xs text-muted-foreground">{label}</label>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          "mt-1 flex items-center justify-between w-full h-9 px-3 py-1 text-sm rounded-md border border-input bg-background",
          isOpen && "ring-1 ring-ring"
        )}
      >
        <span className="flex items-center gap-2">
          <Filter className="h-3.5 w-3.5 text-muted-foreground" />
          {selectedOption?.label || placeholder}
          {selectedOption?.count !== undefined && (
            <span className="text-muted-foreground">({selectedOption.count})</span>
          )}
        </span>
        <ChevronDown className={cn("h-4 w-4 transition-transform", isOpen && "rotate-180")} />
      </button>

      {isOpen && (
        <div className="absolute z-50 mt-1 w-full rounded-md border bg-popover shadow-md max-h-60 overflow-auto">
          <div className="p-1">
            <button
              type="button"
              onClick={() => {
                onChange("");
                setIsOpen(false);
              }}
              className={cn(
                "w-full text-left px-3 py-2 text-sm rounded hover:bg-accent",
                !value && "bg-accent"
              )}
            >
              {placeholder}
            </button>
            {options.map((option) => (
              <button
                key={option.value}
                type="button"
                onClick={() => {
                  onChange(option.value);
                  setIsOpen(false);
                }}
                className={cn(
                  "w-full text-left px-3 py-2 text-sm rounded hover:bg-accent flex items-center justify-between",
                  value === option.value && "bg-accent"
                )}
              >
                {option.label}
                {option.count !== undefined && (
                  <span className="text-muted-foreground text-xs">({option.count})</span>
                )}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

interface FilterBarProps {
  searchValue: string;
  onSearchChange: (value: string) => void;
  filters?: Array<{
    key: string;
    label: string;
    options: FilterOption[];
    value: string;
    onChange: (value: string) => void;
  }>;
  onClearFilters?: () => void;
  className?: string;
}

function FilterBar({ searchValue, onSearchChange, filters, onClearFilters, className }: FilterBarProps) {
  const hasActiveFilters = filters?.some((f) => f.value !== "") || searchValue !== "";

  return (
    <div className={cn("flex flex-col sm:flex-row gap-4", className)}>
      <div className="flex-1">
        <SearchInput value={searchValue} onChange={onSearchChange} />
      </div>

      {filters && filters.length > 0 && (
        <div className="flex flex-wrap gap-4">
          {filters.map((filter) => (
            <FilterSelect
              key={filter.key}
              label={filter.label}
              options={filter.options}
              value={filter.value}
              onChange={filter.onChange}
            />
          ))}
        </div>
      )}

      {hasActiveFilters && onClearFilters && (
        <Button variant="ghost" size="sm" onClick={onClearFilters}>
          <X className="h-4 w-4 mr-1" />
          Limpar
        </Button>
      )}
    </div>
  );
}

interface AdvancedFiltersProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  children: React.ReactNode;
  className?: string;
}

function AdvancedFilters({ open, onOpenChange, children, className }: AdvancedFiltersProps) {
  return (
    <div className={cn("space-y-4", className)}>
      <Button
        variant="ghost"
        size="sm"
        onClick={() => onOpenChange(!open)}
        className="flex items-center gap-2"
      >
        <Filter className="h-4 w-4" />
        Filtros avançados
        <ChevronDown className={cn("h-4 w-4 transition-transform", open && "rotate-180")} />
      </Button>

      {open && (
        <div className="p-4 border rounded-lg bg-muted/20 space-y-4">
          {children}
        </div>
      )}
    </div>
  );
}

export {
  Input,
  SearchInput,
  FilterSelect,
  FilterBar,
  AdvancedFilters,
};
