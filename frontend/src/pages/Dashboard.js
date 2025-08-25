import React from 'react';
import { signOut } from "firebase/auth";
import { auth } from '../firebase';

function Dashboard() {

  const handleLogout = () => {
    signOut(auth).then(() => {
      // Sign-out successful.
    }).catch((error) => {
      // An error happened.
    });
  }

  return (
    <div>
      <h1>Dashboard</h1>
      <button onClick={handleLogout}>Logout</button>
      {/* Add dashboard content here */}
    </div>
  );
}

export default Dashboard;
