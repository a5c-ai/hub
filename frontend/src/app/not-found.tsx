import Link from 'next/link';
import { Button } from '@/components/ui/Button';

export default function NotFound() {
  return (
    <div className="min-h-screen bg-background flex items-center justify-center px-4">
      <div className="max-w-md w-full text-center">
        <h1 className="text-2xl font-bold text-foreground mb-4">Page Not Found</h1>
        <p className="text-muted-foreground mb-6">
          The page you&apos;re looking for does not exist.
        </p>
        <Button asChild variant="default">
          <Link href="/">Go to home</Link>
        </Button>
      </div>
    </div>
  );
}
