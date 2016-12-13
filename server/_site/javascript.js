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
};

const transformSelect = () => {
  return React.createElement(
    'div',
    { className: 'inputGroup form-group' },
    React.createElement('label',
      {
        htmlFor: fieldName,
        className: ''
      },
      "Type of Transform:"
    ),
    React.createElement('select', {
      name: 'type'
    },
      React.createElement('option',
        null,
        "BodyToHeaderTransform"
      ),
      React.createElement('option',
        null,
        "ConstantTransform"
      ),
      React.createElement('option',
        null,
        "HeaderToHeaderTransform"
      ),
      React.createElement('option',
        null,
        "HeaderInjectionTransform"
      )
    )
  )
};
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
};

const CreateArchiveForm = ({ addArchive }) => {
  // Input tracker
  let name, description, source, form;

  form = React.createElement('fieldset',
      { className: 'archive-form' },
      inputGroup('input', 'name', node => { name = node }),
      inputGroup('input', 'description', node => { description = node }),
      inputGroup('textarea', 'source', node => { source = node }),
      React.createElement('legend',
        null,
        'Add a new archive'
      ),
      React.createElement('button',
        {
          onClick: () => {
            addArchive(name, description, source);
          }
        },
        'Save HAR as Traffic Archive'
      )
    )
  return React.createElement('div',
    { className: 'archive-form-wrapper' },
    form,
    React.createElement('button',
      {
        className: 'show-archive-form',
        onClick: function onClick() {
          // WTF please help
          document.getElementsByClassName("archive-form")[0].style.display = 'block',
          document.getElementsByClassName("show-archive-form")[0].style.display = 'none'
        }
      },
      "Create new HAR"
    )
  )
};

const EditArchiveForm = ({ addArchive }) => {
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


const TransformForm = ({ archiveId, addTransform }) => {
  // Input tracker
  let type;

  return React.createElement(
    'fieldset',
    { className: 'transformForm' },
    inputGroup('select', 'type', node => { type = node }),
    inputGroup('input', 'description', node => { type = node }),
    inputGroup('textarea', 'source', node => { type = node }),
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

const Archive = ({ archive, remove, edit }) => {
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
    React.createElement("div",
      null,
      React.createElement("div",
        { className: "name" },
        archive.name
      ),
      React.createElement("div",
        { className: "edit" },
        React.createElement("a",
          {
            onClick: function onClick() {
              edit(archive.id)
            }
          },
          "edit"
        )
      ),
      React.createElement("div",
        { className: "delete" },
        React.createElement("a",
          {
            onClick: function onClick() {
              remove(archive.id)
            }
          },
          "delete"
        )
      )
    ),
    React.createElement("div",
      { className: "description" },
      archive.description
    )
  );
};

const ArchiveList = ({ archives, remove, edit }) => {
  // Map through the archives
  const archiveNodes = archives.map(archive => {
    return React.createElement(Archive, { key: archive.id, archive: archive, name: archive.name, edit: edit, remove: remove });
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
      // The HAR source is stored as raw JSON in the database so we
      // doubly-encode it over HTTP
      source: JSON.stringify(source.value),
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
  // Add archive handler
  updateArchive(id, name, description, source) {
    // Assemble data
    const archive = {
      id: id.value,
      name: name.value,
      description: description.value,
      // The HAR source is stored as raw JSON in the database so we
      // doubly-encode it over HTTP
      source: JSON.stringify(source.value),
    };
    // Update data
    axios.put(this.apiUrl + '/' + id, archive).then(res => {
      this.state.data.forEach(a, idx => {
        if (archive.id == a.id) {
          this.state.data[idx] = res.data;
        }
      });
      this.setState({ data: this.state.data });
      id.value = ''
      name.value = ''
      description.value = ''
      source.value = ''
    });
  }

  // Show a form for editing the archive source as well as creating transforms
  handleEdit(id) {
    console.log("ready to edit")
    // Show forms??
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
      React.createElement(CreateArchiveForm, { addArchive: this.addArchive.bind(this) }),
      React.createElement(ArchiveList, {
        archives: this.state.data,
        edit: this.handleEdit.bind(this),
        remove: this.handleRemove.bind(this)
      }),
      React.createElement(EditArchiveForm, {
        updateArchive: this.updateArchive.bind(this),
        className: 'edit-archive-form hidden'
      })
    );
  }
}

ReactDOM.render(React.createElement(TrafficApp, null), document.getElementById('container'));
