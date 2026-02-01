import { listTags } from "@/lib/api/tags/list";
import HomeClient from "./HomeClient";

const PAGE_SIZE = 10;

export default async function Home() {
  // SSRで初期データをフェッチ
  const initialData = await listTags({
    page: 1,
    pageSize: PAGE_SIZE,
  });

  return <HomeClient initialData={initialData} />;
}
