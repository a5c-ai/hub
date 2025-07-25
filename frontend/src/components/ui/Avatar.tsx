import * as React from 'react';
import { cn, getInitials } from '@/lib/utils';

interface AvatarProps extends React.HTMLAttributes<HTMLDivElement> {
  src?: string;
  alt?: string;
  name?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl';
  fallback?: string;
}

const sizeClasses = {
  sm: 'h-8 w-8 text-xs',
  md: 'h-10 w-10 text-sm',
  lg: 'h-12 w-12 text-base',
  xl: 'h-16 w-16 text-lg',
  '2xl': 'h-20 w-20 text-xl',
};

export function Avatar({
  src,
  alt,
  name,
  size = 'md',
  fallback,
  className,
  ...props
}: AvatarProps) {
  const [imageError, setImageError] = React.useState(false);
  const [imageLoaded, setImageLoaded] = React.useState(false);

  const displayFallback = React.useMemo(() => {
    if (fallback) return fallback;
    if (name) return getInitials(name);
    return '?';
  }, [fallback, name]);

  React.useEffect(() => {
    setImageError(false);
    setImageLoaded(false);
  }, [src]);

  const handleImageLoad = () => {
    setImageLoaded(true);
  };

  const handleImageError = () => {
    setImageError(true);
  };

  return (
    <div
      className={cn(
        'relative inline-flex shrink-0 overflow-hidden rounded-full bg-muted',
        sizeClasses[size],
        className
      )}
      {...props}
    >
      {src && !imageError ? (
        <>
          <img
            className={cn(
              'aspect-square h-full w-full object-cover transition-opacity',
              imageLoaded ? 'opacity-100' : 'opacity-0'
            )}
            src={src}
            alt={alt || name || 'Avatar'}
            onLoad={handleImageLoad}
            onError={handleImageError}
          />
          {!imageLoaded && (
            <div className="absolute inset-0 flex items-center justify-center bg-muted">
              <div className="h-4 w-4 animate-spin rounded-full border-2 border-border border-t-foreground" />
            </div>
          )}
        </>
      ) : (
        <div className="flex h-full w-full items-center justify-center bg-muted font-medium text-muted-foreground">
          {displayFallback}
        </div>
      )}
    </div>
  );
}

export default Avatar;