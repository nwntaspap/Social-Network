import { render, screen } from '@testing-library/react';
import HomePage from './page';

describe('HomePage', () => {
  it('renders without crashing', () => {
    render(<HomePage />);
    // Just verify the component renders something
    const heading = screen.getByRole('heading', { level: 1 });
    expect(heading).toBeInTheDocument();
  });

  it('displays the welcome message', () => {
    render(<HomePage />);
    expect(screen.getByText(/Welcome to Social Network/i)).toBeInTheDocument();
  });
});
