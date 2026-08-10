package engine

func TypeInfer(name, src string, resolver ClassResolver) (*ClassObject, TypeEnv) {
	cls := NewClassObject(name, ParseClass(src))
	env := NewTypeEnv()
	pipeline := DefaultPipeline()

	lineage := cls.Lineage(resolver)
	parentReturns := map[string]DynType{}
	for _, c := range lineage {
		pipeline.Run(c, env, parentReturns)
		parentReturns = env.SnapshotReturns()
	}
	return cls, env
}
