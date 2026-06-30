import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';

// Mock the fonts before importing the component
vi.mock('next/font/google', () => ({
  Rubik: vi.fn(() => ({
    variable: '--font-rubik',
    className: 'mock-rubik-class',
    style: { fontFamily: 'mock-rubik' },
  })),
  Poppins: vi.fn(() => ({
    variable: '--font-poppins',
    className: 'mock-poppins-class',
    style: { fontFamily: 'mock-poppins' },
  })),
}));

// Import the component after mocking
import RootLayout from './layout';

describe('RootLayout', () => {
  it('renders children correctly', () => {
    render(
      <RootLayout>
        <div>Test Content</div>
      </RootLayout>
    );
    expect(screen.getByText('Test Content')).toBeInTheDocument();
  });

  it('renders without crashing', () => {
    const { container } = render(
      <RootLayout>
        <div>Child Content</div>
      </RootLayout>
    );
    expect(container).toBeInTheDocument();
  });
});
