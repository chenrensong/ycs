// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System.Collections.Generic;
using System.Linq;
using Ycs.Contracts;

namespace Ycs.Types
{

    public class YEvent : IYEvent
    {
        private ChangesCollection _changes = null;

        internal YEvent(IAbstractType target, ITransaction transaction)
        {
            Target = target;
            CurrentTarget = target;
            Transaction = transaction;
        }

        public IAbstractType Target { get; set; }
        public IAbstractType CurrentTarget { get; set; }
        public ITransaction Transaction { get; set; }

        public IReadOnlyCollection<object> Path => GetPathTo(CurrentTarget, Target);
        public ChangesCollection Changes => CollectChanges();

        /// <summary>
        /// Check if a struct is added by this event.
        /// </summary>
        internal bool Deletes(IItem str)
        {
            return Transaction.DeleteSet.IsDeleted(str.Id);
        }

        internal bool Adds(IItem str)
        {
            return !Transaction.BeforeState.TryGetValue(str.Id.Client, out var clock) || str.Id.Clock >= clock;
        }

        private ChangesCollection CollectChanges()
        {
            if (_changes == null)
            {
                var target = Target;
                var added = new HashSet<IItem>();
                var deleted = new HashSet<IItem>();
                var delta = new List<Delta>();
                var keys = new Dictionary<string, ChangeKey>();

                _changes = new ChangesCollection
                {
                    Added = added,
                    Deleted = deleted,
                    Delta = delta,
                    Keys = keys
                };

                if (!Transaction.Changed.TryGetValue(Target, out var changed))
                {
                    changed = new HashSet<string>();
                    Transaction.Changed[Target] = changed;
                }

                if (changed.Contains(null))
                {
                    Delta lastOp = null;

                    void packOp()
                    {
                        if (lastOp != null)
                        {
                            delta.Add(lastOp);
                        }
                    }

                    for (var item = Target.Start; item != null; item = item.Right as IItem)
                    {
                        if (item.Deleted)
                        {
                            if (Deletes(item) && !Adds(item))
                            {
                                if (lastOp == null || lastOp.Delete == null)
                                {
                                    packOp();
                                    lastOp = new Delta { Delete = 0 };
                                }

                                lastOp.Delete += item.Length;
                                deleted.Add(item);
                            }
                            else
                            {
                                // Do nothing.
                            }
                        }
                        else
                        {
                            if (Adds(item))
                            {
                                if (lastOp == null || lastOp.Insert == null)
                                {
                                    packOp();
                                    lastOp = new Delta { Insert = new List<object>(1) };
                                }

                                (lastOp.Insert as List<object>).AddRange(item.Content.GetContent());
                                added.Add(item);
                            }
                            else
                            {
                                if (lastOp == null || lastOp.Retain == null)
                                {
                                    packOp();
                                    lastOp = new Delta { Retain = 0 };
                                }

                                lastOp.Retain += item.Length;
                            }
                        }
                    }

                    if (lastOp != null && lastOp.Retain == null)
                    {
                        packOp();
                    }
                }

                foreach (var key in changed)
                {
                    if (key != null)
                    {
                        ChangeAction action;
                        object oldValue;
                        var item = target.Map[key];

                        if (Adds(item))
                        {
                            var prev = item.Left;
                            while (prev != null && Adds(prev))
                            {
                                prev = (prev as IItem).Left;
                            }

                            if (Deletes(item))
                            {
                                if (prev != null && Deletes(prev))
                                {
                                    action = ChangeAction.Delete;
                                    oldValue = (prev as IItem).Content.GetContent().Last();
                                }
                                else
                                {
                                    break;
                                }
                            }
                            else
                            {
                                if (prev != null && Deletes(prev))
                                {
                                    action = ChangeAction.Update;
                                    oldValue = (prev as IItem).Content.GetContent().Last();
                                }
                                else
                                {
                                    action = ChangeAction.Add;
                                    oldValue = null;
                                }
                            }
                        }
                        else
                        {
                            if (Deletes(item))
                            {
                                action = ChangeAction.Delete;
                                oldValue = item.Content.GetContent().Last();
                            }
                            else
                            {
                                break;
                            }
                        }

                        keys[key] = new ChangeKey { Action = action, OldValue = oldValue };
                    }
                }
            }

            return _changes;
        }

        /// <summary>
        /// Compute the path from this type to the specified target.
        /// </summary>
        private IReadOnlyCollection<object> GetPathTo(IAbstractType parent, IAbstractType child)
        {
            var path = new Stack<object>();

            while (child.Item != null && child != parent)
            {
                if (!string.IsNullOrEmpty(child.Item.ParentSub))
                {
                    // Parent is map-ish.
                    path.Push(child.Item.ParentSub);
                }
                else
                {
                    // Parent is array-ish.
                    int i = 0;
                    IItem c = (child.Item.Parent as IAbstractType).Start;
                    while (c != child.Item && c != null)
                    {
                        if (!c.Deleted)
                        {
                            i++;
                        }

                        c = (c as IItem)?.Right;
                    }

                    path.Push(i);
                }

                child = child.Item.Parent as IAbstractType;
            }

            return path;
        }
    }
}
