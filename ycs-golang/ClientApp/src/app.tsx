import React from 'react';
import { Navbar, Nav, NavItem, NavLink } from 'reactstrap';
import { YjsMonacoEditor } from './components/monacoEditor';

export const App = () => {
  return (
    <div>
      <header>
        <Navbar color="light" light expand="md">
          <Nav navbar>
            <NavItem>
              <NavLink href="/">Monaco Editor</NavLink>
            </NavItem>
          </Nav>
        </Navbar>
      </header>

      <YjsMonacoEditor />
    </div>
  );
};
