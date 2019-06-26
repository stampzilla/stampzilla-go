import React from 'react';
import PropTypes from 'prop-types';
import { Editor as SlateEditor } from 'slate-react';
import { Value } from 'slate';
import { connect } from 'react-redux';
import classnames from 'classnames';

import styles from './editor.scss';

const mapStateToProps = state => ({
  devices: state.getIn(['devices', 'list']),
  nodes: state.getIn(['nodes', 'list']),
});

const format = (value) => {
  switch (typeof value) {
    case 'boolean':
      return (
        <span style={{ color: value ? 'green' : 'red' }}>
          {JSON.stringify(value)}
        </span>
      );
    case 'number':
      return <span style={{ color: 'blue' }}>{value}</span>;
    case 'undefined':
    case 'null':
      return <span style={{ color: 'gray' }}>{typeof value}</span>;
    default:
      return JSON.stringify(value);
  }
};

const DeviceNode = connect(mapStateToProps)((props) => {
  const { devices, nodes } = props;
  const id = props.node.getIn(['data', 'id']);
  const node = nodes.get(id.split('.')[0]);

  const value = devices.getIn([
    id,
    'state',
    props.node.getIn(['data', 'state']),
  ]);

  return (
    <span className={styles.node}>
      <span
        style={{ background: '#ddd', padding: '3px' }}
        {...props.attributes}
      >
        {(node && node.get('name')) || node.get('type')}
        {'/'}
        {devices.getIn([id, 'alias']) || devices.getIn([id, 'name']) || id}
        {'/'}
        {props.node.getIn(['data', 'state'])}
      </span>
      <span style={{ background: '#eee', padding: '3px' }}>
        {format(value)}
      </span>
    </span>
  );
});

const schema = {
  inlines: {
    device: {
      // It's important that we mark the mentions as void nodes so that users
      // can't edit the text of the mention.
      isVoid: true,
    },
  },
};

const deserialize = (value) => {
  const j = {
    document: {
      nodes: [
        {
          object: 'block',
          type: 'paragraph',
          nodes: [],
        },
      ],
    },
  };

  const re = /(devices\["([^"]+)"\]\.(\w+))/gm;
  const blocks = value.split(re);

  j.document.nodes[0].nodes = blocks
    .map((block, index) => {
      if ((index - 2) % 4 === 0 || (index - 3) % 4 === 0) {
        return undefined;
      }
      if (!block) {
        return undefined;
      }
      if (block.substring(0, 7) === 'devices') {
        return {
          object: 'inline',
          type: 'device',
          data: {
            text: block,
            id: blocks[index + 1],
            state: blocks[index + 2],
          },
          text: block,
        };
      }
      return {
        object: 'text',
        text: block,
      };
    })
    .filter(Boolean);

  return Value.fromJSON(j);
};

const serialize = (value) => {
  const nodes = value.getIn(['document', 'nodes', '0', 'nodes']);
  let text = '';

  if (nodes) {
    nodes.forEach((node) => {
      switch (node.get('type')) {
        case 'device':
          text = text.concat(node.getIn(['data', 'text']));
          break;
        default:
          if (node.get('leaves')) {
            node.get('leaves').forEach((leaf) => {
              text = text.concat(leaf.get('text'));
            });
          }
          if (node.get('text')) {
            text = text.concat(node.get('text'));
          }
          break;
      }
    });
  }

  return text;
};

const readFromProps = (props) => {
  const value = deserialize(props.value || '');
  return { value };
};

class Editor extends React.Component {
  editorRef = React.createRef();

  constructor(props) {
    super(props);
    this.state = {
      ...readFromProps(props),
    };
  }

  componentWillReceiveProps(props) {
    this.setState({
      ...readFromProps(props),
    });
  }

  onChange = (data) => {
    const { value } = data;
    this.setState({ value });
  };

  renderNode = (props, editor, next) => {
    const { node } = props;
    switch (node.type) {
      case 'device':
        return <DeviceNode {...props} />;
      default:
        return next();
    }
  };

  render() {
    const {
      // id,
      // value,
      // required,
      disabled,
      readonly,
      // label,
      // autofocus,
      // onChange,
    } = this.props;

    return (
      <div
        className={classnames(
          styles.editor,
          'form-control',
          (disabled || readonly) && 'disabled',
        )}
        style={{
          height: 'auto',
        }}
      >
        <input type="hidden" value={serialize(this.state.value)} id="editor" />
        <SlateEditor
          value={this.state.value}
          onChange={this.onChange}
          renderInline={this.renderNode}
          schema={schema}
          ref={this.editorRef}
          autoCorrect={false}
          spellCheck={false}
        />
      </div>
    );
  }
}

Editor.propTypes = {
  id: PropTypes.string,
  value: PropTypes.string,
  required: PropTypes.bool,
  disabled: PropTypes.bool,
  readonly: PropTypes.bool,
  label: PropTypes.string,
  autofocus: PropTypes.bool,
  onChange: PropTypes.func,
};

export default Editor;
