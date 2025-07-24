import { render, screen } from '@testing-library/react';
import { Avatar } from '../Avatar';

describe('Avatar', () => {
  it('displays initials when no image provided', () => {
    render(<Avatar name="John Doe" />);
    expect(screen.getByText('JD')).toBeInTheDocument();
  });

  it('displays custom fallback text', () => {
    render(<Avatar fallback="X" />);
    expect(screen.getByText('X')).toBeInTheDocument();
  });

  it('renders image when src is provided', () => {
    render(<Avatar src="/avatar.jpg" alt="User avatar" name="John Doe" />);
    expect(screen.getByAltText('User avatar')).toBeInTheDocument();
  });

  it('applies correct size classes', () => {
    const { rerender } = render(<Avatar name="John Doe" size="sm" />);
    expect(screen.getByText('JD').parentElement).toHaveClass('h-8', 'w-8');

    rerender(<Avatar name="John Doe" size="lg" />);
    expect(screen.getByText('JD').parentElement).toHaveClass('h-12', 'w-12');
  });
});