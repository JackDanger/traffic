const Title = () => {
  return React.createElement(
      'div',
      null,
      React.createElement(
        'div',
        null,
        React.createElement(
          'h1',
          null,
          'Traffic'
        )
      )
    )
}

const inputGroup = (inputType, fieldName, refBinding) => {
  return React.createElement(
    'div',
    { className: 'inputGroup form-group' },
    React.createElement('label',
      {
        htmlFor: fieldName,
        className: ''
      },
      fieldName
    ),
    React.createElement(inputType, {
      ref: refBinding,
      name: fieldName,
      className: 'form-control'
    })
  )
}

const ArchiveForm = ({ addArchive }) => {
  // Input tracker
  let name, description, source;

  return React.createElement(
    'fieldset',
    { className: 'archiveForm' },
    inputGroup('input', 'name', node => { name = node }),
    inputGroup('input', 'description', node => { description = node }),
    inputGroup('textarea', 'source', node => { source = node }),
    React.createElement(
      'legend',
      null,
      'Add a new archive'
    ),
    React.createElement(
      'button',
      { onClick: () => {
          addArchive(name, description, source);
        } },
      'Save HAR as Traffic Archive'
    )
  );
};

const Archive = ({ archive, remove }) => {
  // Each Archive
  return React.createElement(
    'div',
    {
      name: archive.name,
      remove: remove,
      className: "archive",
      onClick: () => {
        console.log(archive.id);
      }
    },
    React.createElement(
      "div",
      { className: "name" },
      archive.name
    ),
    React.createElement(
      "div",
      { className: "description" },
      archive.description
    )
  );
};

const ArchiveList = ({ archives, remove }) => {
  // Map through the archives
  const archiveNodes = archives.map(archive => {
    return React.createElement(Archive, { key: archive.id, archive: archive, name: archive.name, remove: remove });
  });
  return React.createElement(
    'fieldset',
    null,
    React.createElement('legend',
      null,
      "Archives"
    ),
    archiveNodes
  );
};

// Contaner Component (Ignore for now)
class TrafficApp extends React.Component {
  constructor(props) {
    // Pass props to parent class
    super(props);
    // Set initial state
    this.state = {
      data: []
    };
    this.apiUrl = '/archives';
  }

  // Lifecycle method
  componentDidMount() {
    // Make HTTP reques with Axios
    axios.get(this.apiUrl).then(res => {
      // Set state with result
      this.setState({ data: res.data });
    });
  }

  // Add archive handler
  addArchive(name, description, source) {
    // Assemble data
    const archive = {
      name: name.value,
      description: description.value,
      source: source.value,
    };
    // Update data
    axios.post(this.apiUrl, archive).then(res => {
      this.state.data.push(res.data);
      this.setState({ data: this.state.data });
      name.value = ''
      description.value = ''
      source.value = ''
    });
  }

  // Handle remove
  handleRemove(id) {
    // Filter all archives except the one to be removed
    const remainder = this.state.data.filter(archive => {
      if (archive.id !== id) return archive;
    });
    // Update state with filter
    axios.delete(this.apiUrl + '/' + id).then(res => {
      this.setState({ data: remainder });
    });
  }

  render() {
    return React.createElement(
      'div',
      null,
      React.createElement(Title, null),
      React.createElement(ArchiveForm, { addArchive: this.addArchive.bind(this) }),
      React.createElement(ArchiveList, {
        archives: this.state.data,
        remove: this.handleRemove.bind(this)
      })
    );
  }
}
ReactDOM.render(React.createElement(TrafficApp, null), document.getElementById('container'));
