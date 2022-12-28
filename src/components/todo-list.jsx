import React from 'react';
import Immutable from 'immutable';
import PropTypes from 'prop-types';

class TodoList extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			editing: null,
			editText: "",
			creating: false
		};
	}

	onStatusChange(item, newStatus) {
		if (this.props.onStatusChange) {
			this.props.onStatusChange(item, newStatus.target.checked);
		}
	}

	onRename(item) {
		if (this.props.onRename) {
			this.props.onRename(item);
		}
	}

	onCreate(item) {
		if (this.props.onCreate) {
			const candidate = item.task.trim();
			if (candidate !== '') {
				this.props.onCreate(item);
			}
		}
	}

	onDelete(item) {
		if (this.props.onDelete) {
			this.props.onDelete(item);
		}
	}

	startEdit(item) {
		this.setState({
			creating: false,
			editing: item.rowid,
			editText: item.task
		});
	}

	onEdit(item, event) {
		const newText = event.target.value;
		const newItem = {
			...item,
			task: newText
		};

		this.setState({
			editing: null
		}, this.onRename.bind(this, newItem));
	}

	updateEditor(event) {
		this.setState({
			editText: event.target.value
		});
	}

	checkAction(item, event) {
		if (event.keyCode === 13) {
			// return key pressed
			if (item === null) {
				const newItem = {
					complete: false,
					task: this.state.editText
				};
				this.stopEdit(this.onCreate.bind(this, newItem));
			} else {
				this.onEdit(item, event);
			}
		}
	}

	stopEdit(cbk) {
		this.setState({
			creating: false,
			editing: null,
			editText: ""
		}, cbk);
	}

	startCreate() {
		this.setState({
			editing: null,
			creating: true,
			editText: ""
		});
	}

	render() {
		return (<ul className="todo-list">
			{this.props.todo.map(item => {
				const jsItem = item.toJS();
				jsItem.lock = this.props.lock.contains(jsItem.rowid);
				jsItem.editing = this.state.editing === jsItem.rowid;
				return jsItem;
			}).map((item, idx) => <li key={item.rowid} className={item.complete ? "deleted": null}>
				<input type="checkbox" onChange={this.onStatusChange.bind(this, item)} checked={item.complete} disabled={item.lock} />
				<span className="task">{ item.editing ? <input type="text" autoFocus value={this.state.editText} onChange={this.updateEditor.bind(this)} onKeyDown={this.checkAction.bind(this, item)} onBlur={this.stopEdit.bind(this, null)} /> : item.task }</span>
				<span className="space"></span>
				<button className="edit inline" onClick={this.startEdit.bind(this, item)}>&#9999;</button>
				<button className="delete inline" onClick={this.onDelete.bind(this, item)} disabled={item.lock}>&times;</button>
			</li>)}
			{this.state.creating ? <li key="new">
				<input type="checkbox" disabled />
				<span className="task"><input type="text" autoFocus value={this.state.editText} onChange={this.updateEditor.bind(this)} onKeyDown={this.checkAction.bind(this, null)} onBlur={this.stopEdit.bind(this, null)} /></span>
			</li> : <li className="add-prompt"><button className="inline create" onClick={this.startCreate.bind(this)}>Create New...</button></li>}
		</ul>);
	}
}

TodoList.propTypes = {
	lock: PropTypes.instanceOf(Immutable.Set),
	todo: PropTypes.instanceOf(Immutable.List),
	onStatusChange: PropTypes.func,
	onRename: PropTypes.func,
	onCreate: PropTypes.func,
	onDelete: PropTypes.func
};

export default TodoList;