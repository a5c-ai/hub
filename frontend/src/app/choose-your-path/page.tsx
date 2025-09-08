import Link from "next/link";
import { Button, Card, CardContent, CardHeader, CardTitle } from "@/components/ui";

export default function ChooseYourPathPage() {
  return (
    <div className="min-h-screen flex items-center justify-center p-6">
      <div className="w-full max-w-3xl">
        <Card>
          <CardHeader>
            <CardTitle>Choose your path</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="p-4 border rounded-md">
                <h3 className="text-lg font-medium mb-2">Add to an existing repo</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  Enable Hub for a repository that already exists.
                </p>
                <Button asChild className="w-full">
                  <Link href="/enable">Add to existing repo</Link>
                </Button>
              </div>
              <div className="p-4 border rounded-md">
                <h3 className="text-lg font-medium mb-2">Create a new repo</h3>
                <p className="text-sm text-muted-foreground mb-4">
                  Start fresh by creating a new repository.
                </p>
                <Button asChild className="w-full">
                  <Link href="/new-repo">Create a new repo</Link>
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export const metadata = {
  title: "Choose Your Path - Hub",
  description: "Get started by adding Hub to an existing repository or creating a new one.",
};
