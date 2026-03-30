import { createBrowserRouter, RouterProvider } from "react-router-dom";
import "./App.css";

import Error from "./pages/error";
import HomePage from "./pages/home";
import Root from "./pages/root";

const router = createBrowserRouter([
  {
    path: "/",
    element: <Root />,
    errorElement: <Error />,
    children: [{ index: true, path: "/", element: <HomePage /> }],
  },
]);

function App() {
  return <RouterProvider router={router} />;
}

export default App;
