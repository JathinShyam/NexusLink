import { useEffect, useRef } from "react";
import { SiteFooter } from "./components/SiteFooter";
import { SiteHeader } from "./components/SiteHeader";
import { useApiStatus } from "./hooks/useApiStatus";
import { useTheme } from "./hooks/useTheme";
import { useLocation } from "./lib/navigation";
import { AuthPage } from "./pages/AuthPage";
import { HomePage } from "./pages/HomePage";
import { SignOutPage } from "./pages/SignOutPage";
import "./App.css";

export default function App() {
  const { theme, toggleTheme } = useTheme();
  const apiStatus = useApiStatus();
  const { pathname } = useLocation();
  const previousPath = useRef(pathname);

  useEffect(() => {
    if (previousPath.current === pathname) {
      return;
    }
    previousPath.current = pathname;

    if (pathname === "/signin" || pathname === "/register") {
      return;
    }

    const frame = window.requestAnimationFrame(() => {
      document.getElementById("main")?.focus({ preventScroll: true });
    });
    return () => window.cancelAnimationFrame(frame);
  }, [pathname]);

  let page;
  if (pathname === "/signin") {
    page = <AuthPage key="signin" mode="signin" />;
  } else if (pathname === "/register") {
    page = <AuthPage key="register" mode="register" />;
  } else if (pathname === "/signout") {
    page = <SignOutPage />;
  } else {
    page = <HomePage />;
  }

  return (
    <div className="app">
      <a className="skip-link" href="#main">
        Skip to content
      </a>

      <SiteHeader theme={theme} onToggleTheme={toggleTheme} />

      {page}

      <SiteFooter status={apiStatus} />
    </div>
  );
}
