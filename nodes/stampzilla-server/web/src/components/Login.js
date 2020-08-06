import React, { Component, createRef } from 'react';
import classnames from 'classnames';

class Login extends Component {
  static propTypes = {};

  state = {
    username: '',
    password: '',
  };

  constructor(props) {
    super(props);
    this.autofocusRef = createRef(null);
  }

  componentDidMount() {
    if (this.autofocusRef.current) {
      this.autofocusRef.current.focus();
    }
  }

  onChange = (field, value) => {
    this.setState({
      [field]: value,
    });
  };

  onSubmit = (e) => {
    e.preventDefault();

    this.props.onSubmit(this.state.username, this.state.password);
  };

  render() {
    const { username, password } = this.state;
    const { error } = this.props;

    return (
      <form onSubmit={this.onSubmit}>
        <div className="form-group">
          <label htmlFor="username">Email</label>
          <input
            type="text"
            className={classnames('form-control', error && 'is-invalid')}
            id="username"
            onChange={(e) => this.onChange('username', e.target.value)}
            onFocus={(e) => e.currentTarget.select()}
            value={username}
            ref={this.autofocusRef}
            autoFocus
          />
        </div>
        <div className="form-group">
          <label htmlFor="password">Password</label>
          <input
            type="password"
            className={classnames('form-control', error && 'is-invalid')}
            id="password"
            onChange={(e) => this.onChange('password', e.target.value)}
            onFocus={(e) => e.currentTarget.select()}
            value={password}
          />
          {error && <div className="invalid-feedback">{error.message}</div>}
        </div>
        <button type="submit" className="btn btn-primary">
          Login
        </button>
      </form>
    );
  }
}

export default Login;
