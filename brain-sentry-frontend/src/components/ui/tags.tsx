import { useState } from "react";
import * as React from "react";
import { X, Tag } from "lucide-react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const tagVariants = cva(
  "inline-flex items-center gap-1 rounded-md border px-2 py-1 text-sm transition-colors",
  {
    variants: {
      variant: {
        default: "bg-background hover:bg-accent",
        primary: "bg-primary text-primary-foreground hover:bg-primary/90",
        success: "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400",
        warning: "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400",
        destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
      },
      size: {
        sm: "text-xs px-1.5 py-0.5",
        default: "text-sm px-2 py-1",
        lg: "text-base px-3 py-1.5",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

export interface TagProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof tagVariants> {
  onRemove?: () => void;
}

export function Tag({ className, variant, size, onRemove, children, ...props }: TagProps) {
  return (
    <div className={cn(tagVariants({ variant, size }), className)} {...props}>
      {children}
      {onRemove && (
        <button
          onClick={onRemove}
          className="ml-1 rounded-full hover:bg-accent focus:outline-none focus:ring-1 focus:ring-ring"
        >
          <X className="h-3 w-3" />
        </button>
      )}
    </div>
  );
}

interface TagsInputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, "value" | "onChange"> {
  value: string[];
  onChange: (value: string[]) => void;
  placeholder?: string;
  maxLength?: number;
  maxTags?: number;
  delimiter?: string;
  suggestions?: string[];
  allowDuplicates?: boolean;
  variant?: "default" | "primary" | "success" | "warning" | "destructive";
  size?: "sm" | "default" | "lg";
  className?: string;
  tagClassName?: string;
}

export function TagsInput({
  value = [],
  onChange,
  placeholder = "Digite e pressione Enter...",
  maxLength = 50,
  maxTags = 20,
  delimiter = ",",
  suggestions = [],
  allowDuplicates = false,
  variant = "default",
  size = "default",
  className,
  tagClassName,
}: TagsInputProps) {
  const [inputValue, setInputValue] = useState("");
  const [showSuggestions, setShowSuggestions] = useState(false);
  const inputRef = React.useRef<HTMLInputElement>(null);
  const containerRef = React.useRef<HTMLDivElement>(null);

  const filteredSuggestions = suggestions.filter(
    (suggestion) =>
      !value.includes(suggestion) &&
      suggestion.toLowerCase().includes(inputValue.toLowerCase())
  );

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    if (newValue.includes(delimiter)) {
      const newTags = newValue
        .split(delimiter)
        .map((t) => t.trim())
        .filter((t) => t.length > 0)
        .slice(0, -1); // Remove the last empty element

      addTags(newTags);
      setInputValue("");
    } else {
      setInputValue(newValue);
      setShowSuggestions(newValue.length > 0);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      addTag(inputValue);
    } else if (e.key === "Backspace" && inputValue === "" && value.length > 0) {
      removeTag(value.length - 1);
    }
  };

  const addTag = (tag: string) => {
    const trimmedTag = tag.trim();
    if (
      trimmedTag.length === 0 ||
      trimmedTag.length > maxLength ||
      (maxTags && value.length >= maxTags) ||
      (!allowDuplicates && value.includes(trimmedTag))
    ) {
      return;
    }

    onChange([...value, trimmedTag]);
    setInputValue("");
    setShowSuggestions(false);
  };

  const addTags = (tags: string[]) => {
    const validTags = tags.filter(
      (tag) =>
        tag.trim().length > 0 &&
        tag.length <= maxLength &&
        (allowDuplicates || !value.includes(tag.trim())) &&
        (maxTags ? value.length + tags.length <= maxTags : true)
    );

    if (validTags.length > 0) {
      onChange([...value, ...validTags]);
    }
  };

  const removeTag = (index: number) => {
    onChange(value.filter((_, i) => i !== index));
  };

  // Click outside to close suggestions
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setShowSuggestions(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  return (
    <div ref={containerRef} className={cn("relative", className)}>
      {/* Tags Display */}
      <div className="flex flex-wrap gap-2 p-2 min-h-[42px] border rounded-md bg-background">
        {value.map((tag, index) => (
          <Tag
            key={index}
            variant={variant}
            size={size}
            className={tagClassName}
            onRemove={() => removeTag(index)}
          >
            <Tag className="h-3 w-3" />
            {tag}
          </Tag>
        ))}

        {/* Input */}
        <input
          ref={inputRef}
          type="text"
          value={inputValue}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          placeholder={value.length === 0 ? placeholder : ""}
          className="flex-1 min-w-[120px] bg-transparent border-none outline-none text-sm focus:ring-0"
          disabled={maxTags ? value.length >= maxTags : false}
        />
      </div>

      {/* Suggestions Dropdown */}
      {showSuggestions && filteredSuggestions.length > 0 && (
        <div className="absolute z-50 mt-1 w-full rounded-md border bg-popover shadow-md max-h-48 overflow-auto">
          <div className="p-1">
            {filteredSuggestions.map((suggestion) => (
              <button
                key={suggestion}
                type="button"
                onClick={() => {
                  addTag(suggestion);
                  inputRef.current?.focus();
                }}
                className="w-full text-left px-3 py-2 text-sm rounded hover:bg-accent"
              >
                {suggestion}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Tag Count */}
      {maxTags && (
        <div className="text-xs text-muted-foreground mt-1">
          {value.length}/{maxTags} tags
        </div>
      )}
    </div>
  );
}

// Read-only tags display
interface ReadOnlyTagsProps {
  tags: string[];
  variant?: "default" | "primary" | "success" | "warning" | "destructive";
  size?: "sm" | "default" | "lg";
  className?: string;
  tagClassName?: string;
  maxDisplay?: number;
}

export function ReadOnlyTags({
  tags,
  variant = "default",
  size = "sm",
  className,
  tagClassName,
  maxDisplay = 5,
}: ReadOnlyTagsProps) {
  const displayTags = maxDisplay ? tags.slice(0, maxDisplay) : tags;
  const remainingCount = maxDisplay ? Math.max(0, tags.length - maxDisplay) : 0;

  return (
    <div className={cn("flex flex-wrap gap-1", className)}>
      {displayTags.map((tag, index) => (
        <Tag
          key={index}
          variant={variant}
          size={size}
          className={tagClassName}
        >
          {tag}
        </Tag>
      ))}
      {remainingCount > 0 && (
        <Tag
          variant="default"
          size={size}
          className={tagClassName}
        >
          +{remainingCount} mais
        </Tag>
      )}
    </div  );
}

// Category tag with color mapping
interface CategoryTagProps {
  category: string;
  className?: string;
}

export function CategoryTag({ category, className }: CategoryTagProps) {
  const categoryColors: Record<string, string> = {
    DECISION: "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400",
    PATTERN: "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400",
    ANTIPATTERN: "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400",
    DOMAIN: "bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400",
    BUG: "bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400",
    OPTIMIZATION: "bg-cyan-100 text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-400",
    INTEGRATION: "bg-pink-100 text-pink-700 dark:bg-pink-900/30 dark:text-pink-400",
  };

  const colors = categoryColors[category] || categoryColors.PATTERN;

  return (
    <Tag
      variant="default"
      size="sm"
      className={cn("capitalize font-medium", colors, className)}
    >
      {category.toLowerCase()}
    </Tag>
  );
}

// Importance tag with level colors
interface ImportanceTagProps {
  importance: string;
  className?: string;
}

export function ImportanceTag({ importance, className }: ImportanceTagProps) {
  const importanceColors: Record<string, string> = {
    CRITICAL: "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400",
    IMPORTANT: "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400",
    MINOR: "bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400",
  };

  const colors = importanceColors[importance] || importanceColors.MINOR;

  return (
    <Tag
      variant="default"
      size="sm"
      className={cn("capitalize font-medium", colors, className)}
    >
      {importance.toLowerCase()}
    </Tag>
  );
}
