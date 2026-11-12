import type { AnchorHTMLAttributes, MouseEvent } from "react";
import { navigate } from "./navigation";

type AppLinkProps = AnchorHTMLAttributes<HTMLAnchorElement> & {
  to: string;
  replace?: boolean;
};

export function AppLink({ to, replace, onClick, ...props }: AppLinkProps) {
  function handleClick(event: MouseEvent<HTMLAnchorElement>) {
    onClick?.(event);

    if (
      event.defaultPrevented ||
      event.button !== 0 ||
      event.metaKey ||
      event.ctrlKey ||
      event.shiftKey ||
      event.altKey ||
      props.target === "_blank" ||
      props.download !== undefined
    ) {
      return;
    }

    event.preventDefault();
    navigate(to, { replace });
  }

  return <a href={to} onClick={handleClick} {...props} />;
}
