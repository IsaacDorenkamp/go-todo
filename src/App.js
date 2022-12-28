import './App.css';

import React from 'react';
import axios from 'axios';
import Immutable from 'immutable';

import TodoList from './components/todo-list.jsx';

// Shamelessly hardcoding this since it's
// a small project.
const BASE_URL = "http://localhost:8080";

class App extends React.Component {
	constructor(props) {
		super(props);

		this.state = {
			todo: null,

			// We will lock items by ID to ensure
			// we do not attempt to perform multiple
			// asynchronous actions on a single element!
			locked: Immutable.Set()
		};
	}

	componentDidMount() {
		axios.get(`${BASE_URL}/todo`).then(response => {
			const data = response.data;
			this.setState({
				todo: Immutable.fromJS(data).toList()
			});
		});
	}

	updateTodo(item, newItem) {
		const merged = {
			...item,
			...(newItem || {})
		};

		const itemId = item.rowid;
		this.setState({
			locked: this.state.locked.add(itemId)
		}, () => {
			// The REST API uses url-encoded form data rather
			// than JSON, so we prepare URLSearchParams

			const urlParams = new URLSearchParams();
			urlParams.append('complete', merged.complete);
			urlParams.append('task', merged.task);

			axios.put(`${BASE_URL}/todo/${itemId}`, urlParams).then(response => {
				const entity = response.data.entity;

				const existing = this.state.todo.findIndex((test) => test.get('rowid') === item.rowid);
				if (existing >= 0) {
					const newItem = {
						...entity,
						rowid: item.rowid
					};

					const newTodo = this.state.todo.set(existing, Immutable.Map(newItem));
					this.setState({
						todo: newTodo,
						locked: this.state.locked.delete(itemId)
					});
				}
			}).catch(e => {
				console.error(e);

				// Unlock item to allow interaction again
				this.setState({
					locked: this.state.locked.delete(itemId)
				});
			});
		});
	}

	deleteTodo(item) {
		const itemId = item.rowid;
		this.setState({
			locked: this.state.locked.add(itemId)
		}, () => {
			axios.delete(`${BASE_URL}/todo/${itemId}`).then(response => {
				// Nothing to do, simply reflect deletion in UI
				this.setState({
					todo: this.state.todo.filter(test => test.get('rowid') !== itemId),
					locked: this.state.locked.delete(itemId)
				});
			}).catch(e => {
				console.error(e);

				this.setState({
					locked: this.state.locked.delete(itemId)
				});
			});
		});
	}

	createTodo(item) {
		const urlParams = new URLSearchParams();
		urlParams.append('complete', item.complete);
		urlParams.append('task', item.task);

		axios.post(`${BASE_URL}/todo`, urlParams).then(response => {
			const newTodo = Immutable.Map(response.data.entity);
			this.setState({
				todo: this.state.todo.push(newTodo)
			});
		}).catch(e => {
			console.error(e);
		});
	}

	render() {
		let content;
		if (this.state.todo !== null) {
			content = <TodoList todo={this.state.todo} lock={this.state.locked} onStatusChange={(item, newStatus) => this.updateTodo(item, { complete: newStatus })} onDelete={this.deleteTodo.bind(this)}
				onRename={this.updateTodo.bind(this)} onCreate={this.createTodo.bind(this)} />
		} else {
			content = <div className="loading">Loading...</div>;
		}

		return (<div className="App">
			<h1>
				To-Do List with Go!
			</h1>
			<hr />
			{content}
		</div>);
	}
}

export default App;
