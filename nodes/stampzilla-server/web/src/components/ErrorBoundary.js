import React from 'react';

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false };
  }

  componentDidCatch(error, info) {
    this.setState({
      hasError: true,
      error,
      info,
    });
  }

  render() {
    const { hasError, error, info } = this.state;

    if (hasError) {
      return (
        <div className="pl-4 pr-4">
          <h1>Something went wrong.</h1>
          <pre>{error.toString()}</pre>

          <strong>Stack trace</strong>
          <pre>{error.stack.replace(/webpack:\/\/\/.\//g, '')}</pre>
          {info && (
            <React.Fragment>
              <strong>Component trace</strong>
              <pre>{info.componentStack.replace(/^\s+|\s+$/g, '')}</pre>
            </React.Fragment>
          )}
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;

export const withBoudary = WrappedComponent => props => (
  <ErrorBoundary>
    <WrappedComponent {...props} />
  </ErrorBoundary>
);
