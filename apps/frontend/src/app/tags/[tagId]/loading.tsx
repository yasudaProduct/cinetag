import { Spinner } from "@/components/ui/spinner";

export default function TagDetailLoading() {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <Spinner size="md" className="text-gray-400" />
    </div>
  );
}
