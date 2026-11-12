import { useEffect } from "react";

const DEFAULT_TITLE = "NexusLink — Long links, cut to size.";

export function usePageTitle(title?: string) {
  useEffect(() => {
    document.title = title ? `${title} — NexusLink` : DEFAULT_TITLE;
  }, [title]);
}
