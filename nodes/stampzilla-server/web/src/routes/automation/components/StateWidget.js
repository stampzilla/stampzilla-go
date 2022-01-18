import { connect } from 'react-redux';
import React, { useState } from 'react';
import StateEditor from './StateEditor';
import { IconButton } from '../../../components/formComponents';

function TextButton(props) {
  // Declare a new state variable, which we'll call "count"
  const [text, setText] = useState('');

  return (
    <div className="m-1">
      <input
        type="text"
        value={text}
        className="m-1"
        onChange={(event) => setText(event.target.value)}
      />
      <IconButton
        icon="plus"
        className="btn-success m-1"
        onClick={() => {
          props.onClick(text);
          setText('');
        }}
      />
    </div>
  );
}

class StateWidgetClass extends React.Component {
  constructor(props) {
    super(props);
    this.state = { formData: { ...props.formData } };
  }

  componentWillReceiveProps(nextProps) {
    this.setState({ formData: { ...nextProps.formData }});
  }
  onChange = () => (dev,key,val) => {
    let newVal = '';
    // special hack to detect type. An alternativt would have been to make the type a select box.
	if ( typeof val === 'string' ){
		if ( val === 'true') {
		  newVal = true;
		}else if ( val === 'false') {
		  newVal = false;
		}else if ( val.charAt(val.length-1) === '.') {
			newVal = val;
		}else if (!Number.isNaN(val) && !Number.isNaN(parseFloat(val))) {
		  newVal = parseFloat(val, 10);
		}else{
			newVal = val;
		}
	}else{
		newVal = val;
	}

	  console.log("type", typeof newVal);
    this.setState(prevState => ({
      formData: {
        ...prevState.formData,
        [dev]: { ...prevState.formData[dev], [key]: newVal}
      }}), () => this.props.onChange(this.state.formData));
  }

  onRemoveKey = (dev,key) => (e) => {
    const state = { ...this.state.formData };
    delete state[dev][key];
    this.setState({formData: state }, () => this.props.onChange(this.state.formData));
  }

  onRemoveDevice = (dev) => (e) => {
    const state = { ...this.state.formData };
    delete state[dev];
    this.setState({ formData: state }, () => this.props.onChange(this.state.formData));
  }

  onAddClick = (dev) => (key) => {
	  if ( key == ''){
		  return; // do nothing on empty key
	  }
    this.setState({
      formData: {
        ...this.state.formData,
        [dev]: { ...this.state.formData[dev], [key]: ''}
      }}, () => this.props.onChange(this.state.formData));
  }

  getNodeByUUID(uuid) {
    const { nodes } = this.props;
    const node = nodes && nodes.find((n) => n.get('uuid') === uuid);
    return node;
  }

  getDeviceByUUID(uuid) {
    const { devices } = this.props;
    const dev = devices && devices.find((n) => n.get('id') === uuid);
    return dev;
  }

  render() {
    //const { devices } = this.props;
    //TODO support float, boolean and strings
    return (
      <>
        <div key="widget_root">
          <div className="font-weight-bold">State</div>
          <div className="d-flex flex-column">
            { this.state.formData && Object.keys(this.state.formData).map((row) => {

              const s = row.split('.');
              const node = this.getNodeByUUID(s[0]);
              const dev = this.getDeviceByUUID(row);
              return (
                <div className="d-flex flex-column" key={row}>
                  <div className="d-flex my-2">
                    <div className="p-2">{node && node.get('name') }</div>
                    <div className="p-2">{(dev && dev.get('name')) || 'unknown device'}</div>
                    <div className="flex-grow-1 d-flex align-items-center">
                      <IconButton
                        type="danger"
                        icon="trash"
                        className="array-item-remove"
                        tabIndex="-1"
                        onClick={this.onRemoveDevice(row)}
                      />
                    </div>
                  </div>
                  <table className="table table-striped table-valign-middle">
          <thead>
          </thead>
          <tbody>
                    { this.state.formData[row] && Object.keys(this.state.formData[row]).map((st) => (
                      <tr className="" key={st}>
                        <td style={{ width: '10%' }} className="">{st}</td>
                        <td style={{ width: '80%' }} className="">
                          <StateEditor
                            //className="ml-4"
                            onChange={this.onChange()}
                            state={this.state.formData[row][st]}
                            arrayKey={st}
                            device={row}
                          />
                        </td>
                        <td style={{ width: '10%' }} className="p-1">
                          <IconButton
                            type="danger"
                            icon="trash"
                            className="array-item-remove"
                            tabIndex="-1"
                            //style={btnStyle}
                            //disabled={props.disabled || props.readonly}
                            onClick={this.onRemoveKey(row,st)}
                          />
                        </td>
                      </tr>
                    ))}
          </tbody>
                  </table>
                  <TextButton
                    onClick={this.onAddClick(row)}
                  />
                  <hr style={{ width: '100%' }} />
                </div>
              );
            })}
          </div>
        </div>
      </>
    );
  }
}

const mapStateToProps = (state) => ({
  nodes: state.getIn(['nodes', 'list']),
  devices: state.getIn(['devices', 'list']),
});

export const ConnectedStateWidget = connect(mapStateToProps)(
  StateWidgetClass,
);
const StateWidget = (props) => <ConnectedStateWidget {...props} />;
export default StateWidget;
