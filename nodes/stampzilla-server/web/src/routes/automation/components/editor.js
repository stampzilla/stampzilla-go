import React from 'react';
import PropTypes from 'prop-types';
import { Editor as SlateEditor } from 'slate-react';
import { Value } from 'slate';
import { connect } from 'react-redux';
import { diff } from 'deep-object-diff';

const mapStateToProps = state => ({
  devices: state.getIn(['devices', 'list']),
  nodes: state.getIn(['nodes', 'list']),
});

const DeviceNode = connect(mapStateToProps)((props) => {
  const { devices, nodes } = props;
  const id = props.node.getIn(['data', 'id']);
  const node = nodes.get(id.split('.')[0]);

  return (
    <React.Fragment>
      <span
        style={{ background: '#ddd', padding: '3px' }}
        {...props.attributes}
      >
        {(node && node.get('name')) || '-'}
        {' -> '}
        {devices.getIn([id, 'name']) || id}
        {' -> '}
        {props.node.getIn(['data', 'state'])}
      </span>
      <span style={{ background: '#eee', padding: '3px' }}>
        {JSON.stringify(
          devices.getIn([id, 'state', props.node.getIn(['data', 'state'])]),
        )}
      </span>
    </React.Fragment>
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
          leaves: [
            {
              text: block,
            },
          ],
        };
      }
      return {
        object: 'text',
        leaves: [
          {
            text: block,
          },
        ],
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
          break;
      }
    });
  }

  return text;
};

class Editor extends React.Component {
  editorRef = React.createRef();

  constructor(props) {
    super(props);
    this.state = {
      ...this.readFromProps(props),
    };
  }

  componentWillMount() {
    console.log('will mounted');
  }

  componentDidMount() {
    console.log('did mount');
  }

  componentWillReceiveProps(props) {
    this.setState({
      ...this.readFromProps(props),
    });
  }

  componentWillUnmount() {
    console.log('will unmount');
  }

  onChange = (data) => {
    const { value } = data;
    console.log('onChange called', diff(value.toJS(), this.state.value.toJS()));

    this.setState({ value }, () => {
      const { onChange } = this.props;

      if (onChange) {
        onChange(serialize(value));
      }
    });
  };

  readFromProps(props) {
    const value = deserialize(props.value || '');

    /* if (
      !this.state
      || props.value !== (this.state && serialize(this.state.value))
    ) { */
    console.log(
      'comp',
      props.value === (this.state && serialize(this.state.value)),
    );

    return { value };
    // }
  }

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
        className={`checkbox custom-control custom-checkbox ${
          disabled || readonly ? 'disabled' : ''
        }`}
      >
        <SlateEditor
          value={this.state.value}
          onChange={this.onChange}
          renderNode={this.renderNode}
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
  value: PropTypes.bool,
  required: PropTypes.bool,
  disabled: PropTypes.bool,
  readonly: PropTypes.bool,
  label: PropTypes.string,
  autofocus: PropTypes.bool,
  onChange: PropTypes.func,
};

export default Editor;
