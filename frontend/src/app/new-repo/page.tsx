import { redirect } from "next/navigation";

export default function NewRepoAliasPage() {
  // Redirect to the existing create repository page
  redirect("/repositories/new");
}

