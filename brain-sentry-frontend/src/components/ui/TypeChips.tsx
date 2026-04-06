interface TypeChipItem {
  label: string;
  count: number;
  color?: string;
}

interface TypeChipsProps {
  items: TypeChipItem[];
  selected?: string | null;
  onSelect: (label: string | null) => void;
  className?: string;
}

export function TypeChips({ items, selected, onSelect, className = "" }: TypeChipsProps) {
  return (
    <div className={`flex flex-wrap gap-1.5 ${className}`}>
      <button
        onClick={() => onSelect(null)}
        className={`px-2.5 py-1 text-xs font-medium rounded-full border transition-colors ${
          !selected
            ? "bg-foreground text-background border-foreground"
            : "bg-transparent text-muted-foreground border-border hover:border-foreground/50"
        }`}
      >
        All
      </button>
      {items.map((item) => (
        <button
          key={item.label}
          onClick={() => onSelect(item.label === selected ? null : item.label)}
          className={`inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-full border transition-colors ${
            selected === item.label
              ? "bg-foreground text-background border-foreground"
              : "bg-transparent text-muted-foreground border-border hover:border-foreground/50"
          }`}
        >
          {item.color && (
            <span className="h-2 w-2 rounded-full flex-shrink-0" style={{ backgroundColor: item.color }} />
          )}
          <span className="capitalize">{item.label.toLowerCase()}</span>
          <span className={`text-[10px] ${selected === item.label ? "text-background/70" : "text-muted-foreground/60"}`}>
            {item.count}
          </span>
        </button>
      ))}
    </div>
  );
}
