import React, { Component } from 'react';

class Login extends Component {
  static propTypes = {};

  state = {
    username: '',
    password: '',
  };

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

    return (
      <form onSubmit={this.onSubmit}>
        <div className="form-group">
          <label htmlFor="username">Username</label>
          <input
            type="text"
            className="form-control"
            id="username"
            onChange={e => this.onChange('username', e.target.value)}
            value={username}
          />
        </div>
        <div className="form-group">
          <label htmlFor="password">Password</label>
          <input
            type="password"
            className="form-control"
            id="password"
            onChange={e => this.onChange('password', e.target.value)}
            value={password}
          />
        </div>
        <button type="submit" className="btn btn-primary">
          Login
        </button>
      </form>
    );
  }
}

export default Login;
