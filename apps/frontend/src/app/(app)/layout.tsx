import { Sidebar } from "../../components/Sidebar";

export default function AppLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <main className="flex-1 md:ml-64 min-h-screen pb-20 md:pb-0">
        {children}
      </main>
    </div>
  );
}
