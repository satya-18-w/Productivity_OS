import type { Category } from "../../api";
import { CategoryIndicator } from "../../components/productivity/CategoryIndicator";
import { IconButton } from "../../components/ui/IconButton";
import { Menu, type MenuItem } from "../../components/ui/Menu";
import { MoreIcon } from "../../components/ui/icons";

export interface CategoryRowProps {
  category: Category;
  onRename: (category: Category) => void;
  onArchive: (category: Category) => void;
}

export function CategoryRow({ category, onRename, onArchive }: CategoryRowProps) {
  const items: MenuItem[] = [
    { key: "rename", label: "Rename", onSelect: () => onRename(category) },
    { key: "sep", separator: true },
    { key: "archive", label: "Archive", danger: true, onSelect: () => onArchive(category) },
  ];

  return (
    <li className="category-row">
      <CategoryIndicator
        variant="tile"
        colorKey={category.id}
        name={category.name}
        glyph={category.name.charAt(0).toUpperCase()}
      />
      <span className="category-row__name">{category.name}</span>
      <Menu
        label={`Actions for ${category.name}`}
        trigger={
          <IconButton label={`Actions for ${category.name}`} size="sm">
            <MoreIcon />
          </IconButton>
        }
        items={items}
      />
    </li>
  );
}
