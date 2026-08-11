import React, { Component } from 'react';
import PropTypes from 'prop-types';
import { Button } from 'reactstrap';
import { connect } from 'react-redux';
import Form from 'react-jsonschema-form';

import { write } from '../../components/Websocket';
import Card from '../../components/Card';
import CustomCheckbox from '../../components/CustomCheckbox';

const INDENT = 2;

// Returns null when the text isn't valid JSON, so callers can leave the
// user's text untouched instead of mangling a work-in-progress edit.
const formatJson = (value) => {
  try {
    return JSON.stringify(JSON.parse(value), null, INDENT);
  } catch (err) {
    return null;
  }
};

// A plain controlled <textarea> instead of a fancy syntax-highlighting
// editor: it has no internal DOM/tokenizer state of its own to desync from
// React, so typing, pasting and invalid JSON can never corrupt the input.
const JsonWidget = (props) => {
  const {
    id, value, onChange, options, disabled, readonly,
  } = props;
  const formatted = formatJson(value || '');

  return (
    <>
      <textarea
        id={id}
        className="form-control"
        style={{ fontFamily: 'monospace' }}
        rows={(options && options.rows) || 15}
        value={value || ''}
        disabled={disabled || readonly}
        onChange={(event) => onChange(event.target.value)}
      />
      <button
        type="button"
        className="btn btn-secondary btn-sm mt-2"
        disabled={formatted === null || disabled || readonly}
        onClick={() => onChange(formatted)}
      >
        Format JSON
      </button>
    </>
  );
};
JsonWidget.propTypes = {
  id: PropTypes.string,
  value: PropTypes.string,
  onChange: PropTypes.func.isRequired,
  options: PropTypes.shape({
    rows: PropTypes.number,
  }),
  disabled: PropTypes.bool,
  readonly: PropTypes.bool,
};
JsonWidget.defaultProps = {
  id: undefined,
  value: '',
  options: {},
  disabled: false,
  readonly: false,
};

const schema = {
  type: 'object',
  required: ['name', 'config'],
  properties: {
    name: {
      type: 'string',
      title: 'Name',
    },
    config: {
      type: 'string',
      title: 'Config',
    },
  },
};
const uiSchema = {
  config: {
    'ui:widget': 'json',
    'ui:options': {
      rows: 15,
    },
  },
};

const validate = (formData, errors) => {
  try {
    JSON.parse(formData.config);
  } catch (err) {
    errors.config.addError(`Invalid JSON: ${err.message}`);
  }
  return errors;
};

class Node extends Component {
  constructor(props) {
    super(props);

    this.state = {
      isValid: true,
      formData: {},
      loadedUuid: null,
    };
  }

  componentDidMount = () => {
    this.syncFormData(this.props);
  };

  componentDidUpdate = () => {
    this.syncFormData(this.props);
  };

  // Only load formData from the store when we don't have it yet for this
  // node. The `nodes` list is replaced with a new reference on every
  // websocket update (e.g. unrelated node heartbeats), so resyncing
  // unconditionally here would wipe out unsaved edits while the user types.
  syncFormData = (props) => {
    const { nodes, match } = props;
    if (this.state.loadedUuid === match.params.uuid) {
      return;
    }
    const node = nodes && nodes.find((n) => n.get('uuid') === match.params.uuid);
    if (node) {
      this.setState({
        loadedUuid: match.params.uuid,
        formData: {
          name: node.get('name'),
          config: JSON.stringify(node.get('config'), null, INDENT),
        },
      });
    }
  };

  onChange = () => (data) => {
    const { errors, formData } = data;
    this.setState({
      isValid: errors.length === 0,
      formData,
    });
  };

  onSubmit = () => ({ formData }) => {
    const { nodes, match } = this.props;
    const node = nodes.find((n) => n.get('uuid') === match.params.uuid);

    write({
      type: 'setup-node',
      body: {
        ...node.toJS(),
        ...formData,
        config: JSON.parse(formData.config),
      },
    });
  };

  onClickNode = (uuid) => () => {
    const { history } = this.props;
    history.push(`/nodes/${uuid}`);
  };

  render() {
    const { nodes, match } = this.props;
    const node = nodes.find((n) => n.get('uuid') === match.params.uuid);

    return (
      <>
        <div className="row">
          <div className="col-md-12">
            <Card
              title={
                node ? (
                  <>
                    Settings 1 for node
                    {' '}
                    <strong>{node.get('uuid')}</strong>
                    {' '}
                    (type
                    {' '}
                    <strong>{node.get('type')}</strong>
                    )
                  </>
                ) : (
                  'Settings'
                )
              }
              bodyClassName="p-0"
            >
              <div className="card-body">
                <Form
                  key={this.state.loadedUuid}
                  schema={schema}
                  uiSchema={uiSchema}
                  showErrorList={false}
                  liveValidate
                  validate={validate}
                  onChange={this.onChange()}
                  formData={this.state.formData}
                  onSubmit={this.onSubmit()}
                  // onError={log('errors')}
                  // disabled={this.props.disabled}
                  // transformErrors={this.props.transformErrors}
                  widgets={{
                    CheckboxWidget: CustomCheckbox,
                    json: JsonWidget,
                  }}
                >
                  <button
                    ref={(btn) => {
                      this.submitButton = btn;
                    }}
                    style={{ display: 'none' }}
                    type="submit"
                  />
                </Form>
              </div>
              <div className="card-footer">
                <Button
                  color="primary"
                  disabled={!this.state.isValid || this.props.disabled}
                  onClick={() => this.submitButton.click()}
                >
                  Save
                </Button>
              </div>
            </Card>
          </div>
        </div>
      </>
    );
  }
}

const mapToProps = (state) => ({
  nodes: state.getIn(['nodes', 'list']),
  connections: state.getIn(['connections', 'list']),
});

export default connect(mapToProps)(Node);
