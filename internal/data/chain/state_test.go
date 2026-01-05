package chain

import (
	"testing"
)

func TestNodeDFA_Transitions(t *testing.T) {
	dfa := NewNodeDFA()
	// initial is Single, ReaderConfirmer
	// set role to Relay so successor connect is allowed
	if err := dfa.Emit(RoleRelay); err != nil {
		t.Fatalf("unexpected err setting role: %v", err)
	}
	<-dfa.States()

	// SuccessorConnect should make it Head
	if err := dfa.Emit(SuccessorConnect); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	st := <-dfa.States()
	if st.Position != Head {
		t.Fatalf("expected Head got %v", st.Position)
	}

	// PredecessorConnect should now move Head -> Middle
	if err := dfa.Emit(PredecessorConnect); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	st2 := <-dfa.States()
	if st2.Position != Middle {
		t.Fatalf("expected Middle got %v", st2.Position)
	}
}

func TestNodeDFA_IllegalTransitions(t *testing.T) {
	// initial Role is ReaderConfirmer; PredecessorConnect should be illegal
	dfa1 := NewNodeDFA()
	if err := dfa1.Emit(PredecessorConnect); err == nil {
		t.Fatalf("expected illegal transition for ReaderConfirmer + PredecessorConnect")
	}
	// and SuccessorConnect should also be illegal for ReaderConfirmer
	if err := dfa1.Emit(SuccessorConnect); err == nil {
		t.Fatalf("expected illegal transition for ReaderConfirmer + SuccessorConnect")
	}

	// a node in Reader role should also reject PredecessorConnect
	dfa2 := NewNodeDFA()
	if err := dfa2.Emit(RoleReader); err != nil {
		t.Fatalf("unexpected err setting RoleReader: %v", err)
	}
	<-dfa2.States()
	if err := dfa2.Emit(PredecessorConnect); err == nil {
		t.Fatalf("expected illegal transition for Reader + PredecessorConnect")
	}

	// a node in Confirmer role should reject SuccessorConnect
	dfa3 := NewNodeDFA()
	if err := dfa3.Emit(RoleConfirmer); err != nil {
		t.Fatalf("unexpected err setting RoleConfirmer: %v", err)
	}
	<-dfa3.States()
	if err := dfa3.Emit(SuccessorConnect); err == nil {
		t.Fatalf("expected illegal transition for Confirmer + SuccessorConnect")
	}
}
