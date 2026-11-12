import { Component, type ErrorInfo, type ReactNode } from "react";
import { RouteMark } from "./Icons";
import "./ErrorBoundary.css";

type Props = {
  children: ReactNode;
};

type State = {
  hasError: boolean;
};

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("NexusLink failed to render", error, info);
  }

  render() {
    if (this.state.hasError) {
      return (
        <main className="error-page">
          <div className="error-page__inner">
            <p className="error-page__brand">
              <RouteMark className="error-page__mark" />
              <span translate="no">NexusLink</span>
            </p>
            <h1 className="error-page__title">Something broke.</h1>
            <p className="error-page__copy">
              Reload the page to try again. If it keeps happening, the API may be
              unavailable.
            </p>
            <button
              className="btn btn--primary"
              type="button"
              onClick={() => window.location.reload()}
            >
              Reload page
            </button>
          </div>
        </main>
      );
    }

    return this.props.children;
  }
}
