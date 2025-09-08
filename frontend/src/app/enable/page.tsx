export default function EnablePage() {
  return (
    <div className="min-h-screen flex items-center justify-center p-6">
      <div className="max-w-xl w-full text-center space-y-4">
        <h1 className="text-2xl font-bold">Enable on an existing repository</h1>
        <p className="text-muted-foreground">
          Select an existing repository to enable Hub features. If you intended to create a repository instead, go to
          {" "}
          <a className="text-primary underline" href="/new-repo">/new-repo</a>.
        </p>
      </div>
    </div>
  );
}

export const metadata = {
  title: "Enable Repository - Hub",
  description: "Enable Hub features on an existing repository.",
};

