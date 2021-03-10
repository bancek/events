package events_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/xerrors"

	. "github.com/bancek/events"
)

type pointerError struct {
	nonComparable []byte
}

func (e pointerError) Error() string {
	return "pointer error"
}

type nonPointerError struct {
	nonComparable []byte
}

func (e nonPointerError) Error() string {
	return "non pointer error"
}

var _ = Describe("Errors", func() {
	Describe("UnwrapAll", func() {
		It("should return the same error", func() {
			err := nonPointerError{}
			unwrapped := UnwrapAll(err)
			Expect(unwrapped).To(Equal(err))
		})

		It("should unwrap a pointer error", func() {
			err := &pointerError{}
			wrapped := xerrors.Errorf("wrap 2: %w", xerrors.Errorf("wrap 1: %w", err))
			unwrapped := UnwrapAll(wrapped)
			Expect(unwrapped).To(Equal(err))
		})

		It("should unwrap a non-pointer error", func() {
			err := nonPointerError{}
			wrapped := xerrors.Errorf("wrap 2: %w", xerrors.Errorf("wrap 1: %w", err))
			unwrapped := UnwrapAll(wrapped)
			Expect(unwrapped).To(Equal(err))
		})

		It("should return nil", func() {
			unwrapped := UnwrapAll(nil)
			Expect(unwrapped).To(BeNil())
		})
	})

	Describe("GetCause", func() {
		It("should return false if the error is not wrapped", func() {
			err := nonPointerError{}
			_, ok := GetCause(err)
			Expect(ok).To(BeFalse())
		})

		It("should unwrap a pointer error", func() {
			err := &pointerError{}
			wrapped := xerrors.Errorf("wrap 2: %w", xerrors.Errorf("wrap 1: %w", err))
			cause, ok := GetCause(wrapped)
			Expect(ok).To(BeTrue())
			Expect(cause).To(Equal(err))
		})

		It("should unwrap a non-pointer error", func() {
			err := nonPointerError{}
			wrapped := xerrors.Errorf("wrap 2: %w", xerrors.Errorf("wrap 1: %w", err))
			cause, ok := GetCause(wrapped)
			Expect(ok).To(BeTrue())
			Expect(cause).To(Equal(err))
		})

		It("should return false if for nil error", func() {
			_, ok := GetCause(nil)
			Expect(ok).To(BeFalse())
		})
	})
})
