import { render, screen } from '@testing-library/react';
import { Input } from '../Input';

describe('Input', () => {
  it('renders input with label', () => {
    render(<Input label="Username" />);
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument();
  });

  it('displays error message', () => {
    render(<Input label="Email" error="Email is required" />);
    expect(screen.getByText('Email is required')).toBeInTheDocument();
    expect(screen.getByLabelText(/email/i)).toHaveClass('border-destructive');
  });

  it('supports different input types', () => {
    const { rerender } = render(<Input type="email" label="Email" />);
    expect(screen.getByLabelText(/email/i)).toHaveAttribute('type', 'email');

    rerender(<Input type="password" label="Password" />);
    expect(screen.getByLabelText(/password/i)).toHaveAttribute('type', 'password');
  });

  it('can be disabled', () => {
    render(<Input label="Username" disabled />);
    expect(screen.getByLabelText(/username/i)).toBeDisabled();
  });
});