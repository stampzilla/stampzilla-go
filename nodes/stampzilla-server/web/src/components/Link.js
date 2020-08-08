import { Link as RouterLink } from 'react-router-dom';
import { withRouter } from 'react-router';
import React from 'react';
import Url from 'url';
import classnames from 'classnames';

class Link extends React.Component {
  constructor() {
    super();
    this.state = {
      isLocal: false,
    };
  }

  /* eslint-disable react/no-did-mount-set-state */
  componentDidMount() {
    const { to } = this.props;
    if (to && typeof window !== 'undefined') {
      const url = Url.parse(to);
      if (
        window.location.hostname === url.hostname
        || !url.hostname
        || !url.hostname.length
      ) {
        this.setState({
          isLocal: true,
          localTo: to.replace('www.', '').replace(window.location.origin, ''),
        });
      }
    }
  }
  /* eslint-enable */

  render() {
    const {
      to,
      className,
      children,
      location,
      activeClass,
      onClick,
      exact,
    } = this.props;
    const { isLocal, localTo } = this.state;
    const active = localTo
      && activeClass
      && ((location.pathname.substring(0, localTo.length) === localTo && !exact)
        || (location.pathname === localTo && exact))
      ? activeClass
      : null;

    return isLocal ? (
      <RouterLink
        to={localTo}
        className={classnames([className, active])}
        onClick={onClick}
      >
        {children}
      </RouterLink>
    ) : (
      <a href={to} className={className} onClick={onClick}>
        {children}
      </a>
    );
  }
}

export default withRouter(Link);
