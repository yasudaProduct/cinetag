import { Spinner } from "@/components/ui/spinner";

export default function RootLoading() {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <Spinner size="md" className="text-gray-400" />
    </div>
  );
}
