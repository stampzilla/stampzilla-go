import React, { Component } from 'react';
import { Button } from 'reactstrap';
import { connect } from 'react-redux';
import Form from 'react-jsonschema-form';
import Select from 'react-select';

import StateWidget from './components/StateWidget';
import { add, save, remove } from '../../ducks/savedstates';
import Card from '../../components/Card';

const schema = {
  type: 'object',
  required: ['name'],
  properties: {
    name: {
      type: 'string',
      title: 'Name',
    },
    state: {
      type: 'object',
      title: 'State',
    },
  },
};
const uiSchema = {
  state: {
    'ui:field': 'StateWidget',
  },
};

class Savedstate extends Component {
  constructor(props) {
    super();

    const { savedstates, match } = props;
    const savedstate = savedstates.find((n) => n.get('uuid') === match.params.uuid);
    const formData = savedstate && savedstate.toJS();
    this.state = {
      formData,
      isValid: true,
    };
  }

  componentWillReceiveProps(nextProps) {
    const { savedstates, match } = nextProps;
    if (
      !this.props
      || match.params.uuid !== this.props.match.params.uuid
      || savedstates !== this.props.savedstates
    ) {
      const savedstate = savedstates.find((n) => n.get('uuid') === match.params.uuid);
      const formData = savedstate && savedstate.toJS();
      if (formData && formData.state == null) {
        formData.state = {};
      }
      this.setState({
        formData,
      });
    }
  }

  onBackClick = () => {
    const { history } = this.props;
    history.push('/aut');
  };

  onChange = () => (data) => {
    const { errors, formData } = data;
    this.setState({
      isValid: errors.length === 0,
      formData,
    });
  };

  onRemove = () => {
    if (confirm('Are you sure?')) {
      const { match, dispatch } = this.props;
      dispatch(remove(match.params.uuid));
      this.onBackClick();
    }
  };


  onSubmit = () => ({ formData }) => {
    const { dispatch } = this.props;

    if (formData.uuid) {
      dispatch(save(formData));
    } else {
      dispatch(add(formData));
    }

    const { history } = this.props;
    history.push('/aut');
  };

  getDevicesArray() {
    const { devices } = this.props;
    const devs = devices.toJS();
    return devs && Object.keys(devs).map((key) => ({ value: devs[key].id, label: devs[key].name }));
  }

  onAddDevice = () => () => {
    if (!this.state.newDeviceId) {
      return;
    }

    this.setState((prevState) => ({ formData: {
      ...prevState.formData,
      state: { ...prevState.formData.state, [prevState.newDeviceId]: {} },
    } }));
  }

  render() {
    const { match } = this.props;

    return (
      <>
        <div className="row">
          <div className="col-md-12">
            <Card
              title={match.params.uuid ? 'Edit scene' : 'New Scene'}
              bodyClassName="p-0"
            >
              <div className="card-body">
                <Form
                  schema={schema}
                  uiSchema={uiSchema}
                  showErrorList={false}
                  liveValidate
                  onChange={this.onChange()}
                  formData={this.state.formData}
                  onSubmit={this.onSubmit()}
                  // onError={log('errors')}
                  // disabled={this.props.disabled}
                  // transformErrors={this.props.transformErrors}
                  // ObjectFieldTemplate={ObjectFieldTemplate}
                  // ArrayFieldTemplate={ArrayFieldTemplate}
                  fields={{
                    StateWidget,
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

                <div className="container pt-5 mx-0">
                  <div className="row">
                    <div className="col">
                      <Select
                        options={this.getDevicesArray()}
                        onChange={(e) => {
                          this.setState({ newDeviceId: e.value });
                        }}

                      />
                    </div>
                    <div className="col">
                      <Button
                        color="primary"
                        onClick={this.onAddDevice()}
                      >
                        Add
                      </Button>
                    </div>
                  </div>
                </div>

              </div>
              <div className="card-footer">
                <Button color="secondary" onClick={this.onBackClick}>
                  Back
                </Button>
                <Button
                  color="danger"
                  disabled={this.props.disabled}
                  onClick={this.onRemove}
                  className="ml-2 btn-sm"
                >
                  Remove
                </Button>
                <Button
                  color="primary"
                  disabled={!this.state.isValid || this.props.disabled}
                  onClick={() => this.submitButton.click()}
                  className="float-right"
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
  devices: state.getIn(['devices', 'list']),
  savedstates: state.getIn(['savedstates', 'list']),
});

export default connect(mapToProps)(Savedstate);
