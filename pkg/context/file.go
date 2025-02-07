package context

// type LocalFileContext struct {
// 	Paths       []string
// 	Directories []string
// }

// func (c *ContextManager) FromFiles(ctx context.Context, localContext caggregates.LocalFileContext, options shared.ContextOptions) error {
// 	files := []string{}
// 	messages := []shared.Message{}
// 	for _, path := range localContext.Files {
// 		_, err := os.Stat(path)
// 		if err != nil {
// 			return err
// 		}
// 		files = append(files, path)
// 	}
// 	for _, directory := range localContext.Directories {
// 		if !directory.Recursive {
// 			f, err := os.ReadDir(directory.Path)
// 			if err != nil {
// 				return err
// 			}
// 			for _, file := range f {
// 				if !file.IsDir() {
// 					files = append(files, filepath.Join(directory.Path, file.Name()))
// 				}
// 			}
// 		} else {
// 			err := filepath.WalkDir(
// 				directory.Path,
// 				func(path string, d fs.DirEntry, err error) error {
// 					if err != nil {
// 						return err
// 					}
// 					if !d.IsDir() {
// 						files = append(files, filepath.Join(path, d.Name()))
// 					}
// 					return nil
// 				})
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	for _, file := range files {
// 		data, err := os.ReadFile(file)
// 		if err != nil {
// 			return err
// 		}
// 		text := fmt.Sprintf("\nfile %s:\n```\n%s\n```", file, data)
// 		id, err := uuid.NewV6()
// 		if err != nil {
// 			return err
// 		}
// 		messages = append(messages, shared.Message{
// 			ID:        id.String(),
// 			Role:      shared.UserRole,
// 			Content:   text,
// 			CreatedAt: time.Now().UTC(),
// 		})
// 	}
// 	id, err := uuid.NewV6()
// 	if err != nil {
// 		return err
// 	}
// 	context := shared.Context{
// 		ID:          id.String(),
// 		Name:        options.Name,
// 		Description: options.Description,
// 		Messages:    messages,
// 		CreatedAt:   time.Now().UTC(),
// 	}

// 	err = c.store.CreateContext(ctx, context)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
