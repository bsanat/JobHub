interface UserAvatarProps {
  fullName: string;
  size?: "sm" | "md" | "lg";
}

export function UserAvatar({ fullName, size = "md" }: UserAvatarProps) {
  const initials = fullName
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toLocaleUpperCase())
    .join("");

  return (
    <span className={`user-avatar user-avatar--${size}`} aria-hidden="true">
      {initials}
    </span>
  );
}
