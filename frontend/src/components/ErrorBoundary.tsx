import { Component, type ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
}

export default class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="page">
          <div className="message error">
            页面出现了意外错误
            <br />
            <button
              className="header-btn"
              style={{ marginTop: 16 }}
              onClick={() => {
                this.setState({ hasError: false });
                window.location.href = '/';
              }}
            >
              返回首页
            </button>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
