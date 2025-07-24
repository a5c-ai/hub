'use client';

import { Label } from '@/types';

interface LabelBadgeProps {
  label: Label;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function LabelBadge({ label, size = 'sm', className = '' }: LabelBadgeProps) {
  const sizeClasses = {
    sm: 'text-xs px-2 py-0.5',
    md: 'text-sm px-2.5 py-1',
    lg: 'text-base px-3 py-1.5',
  };

  // Calculate text color based on background color
  const getTextColor = (bgColor: string) => {
    // Remove # if present
    const color = bgColor.replace('#', '');
    const r = parseInt(color.substr(0, 2), 16);
    const g = parseInt(color.substr(2, 2), 16);
    const b = parseInt(color.substr(4, 2), 16);
    
    // Calculate brightness using relative luminance formula
    const brightness = (r * 299 + g * 587 + b * 114) / 1000;
    return brightness > 128 ? '#000000' : '#ffffff';
  };

  return (
    <span
      className={`inline-flex items-center rounded-full font-medium ${sizeClasses[size]} ${className}`}
      style={{
        backgroundColor: label.color,
        color: getTextColor(label.color),
      }}
      title={label.description}
    >
      {label.name}
    </span>
  );
}