// ------------------------------------------------------------------------------
//  <copyright company="Microsoft Corporation">
//      Copyright (c) Microsoft Corporation.  All rights reserved.
//  </copyright>
// ------------------------------------------------------------------------------

using System;
using System.Collections.Generic;
using System.Diagnostics;
using Ycs.Contracts;
using Ycs.Lib0;
using Ycs.Types;
using Ycs.Utils;

namespace Ycs.Structs
{
    public class Item : IItem
    {
        [Flags]
        private enum InfoEnum : int
        {
            Keep = 1 << 0,
            Countable = 1 << 1,
            Deleted = 1 << 2,
            Marker = 1 << 3
        }

        private InfoEnum _info;

        public Item(StructID id, IItem left, StructID? leftOrigin, IItem right, StructID? rightOrigin, object parent, string parentSub, IContent content)
        {
            this.Id = id;
            this.Length = content.Length;

            LeftOrigin = leftOrigin;
            Left = left;
            Right = right;
            RightOrigin = rightOrigin;
            Parent = parent;
            ParentSub = parentSub;
            Redone = null;
            Content = (IContentEx)content;

            _info = content.Countable ? InfoEnum.Countable : 0;
        }

        /// <summary>
        /// The item that was originally to the left of this item.
        /// </summary>
        public StructID? LeftOrigin { get; set; }

        /// <summary>
        /// The item that is currently to the left of this item.
        /// </summary>
        public IItem Left { get; set; }

        /// <summary>
        /// The item that was originally to the right of this item.
        /// </summary>
        public StructID? RightOrigin { get; set; }

        /// <summary>
        /// The item that is currently to the right of this item.
        /// </summary>
        public IItem Right { get; set; }

        /// <summary>
        /// AbstractType or ID.
        /// </summary>
        public object Parent { get; set; }

        /// <summary>
        /// If the parent refers to this item with some kind of key (e.g. YMap).
        /// The key is then used to refer to the list in which to insert this item.
        /// If 'parentSub = null', type._start is the list in which to insert to.
        /// Otherwise, it is 'parent._map'.
        /// </summary>
        public string ParentSub { get; set; }

        /// <summary>
        /// Refers to the type that undid this operation.
        /// </summary>
        public StructID? Redone { get; set; }

        public IContentEx Content { get; set; }

        public bool Marker
        {
            get => _info.HasFlag(InfoEnum.Marker);
            set
            {
                if (value)
                {
                    _info |= InfoEnum.Marker;
                }
                else
                {
                    _info &= ~InfoEnum.Marker;
                }
            }
        }

        /// <summary>
        /// If set to 'true', do not garbage collect this item.
        /// </summary>
        public bool Keep
        {
            get => _info.HasFlag(InfoEnum.Keep);
            set
            {
                if (value)
                {
                    _info |= InfoEnum.Keep;
                }
                else
                {
                    _info &= ~InfoEnum.Keep;
                }
            }
        }

        public bool Countable
        {
            get => _info.HasFlag(InfoEnum.Countable);
            set
            {
                if (value)
                {
                    _info |= InfoEnum.Countable;
                }
                else
                {
                    _info &= ~InfoEnum.Countable;
                }
            }
        }
        public StructID Id { get; set; }
        public int Length { get; set; }

        public bool Deleted => _info.HasFlag(InfoEnum.Deleted);

        /// <summary>
        /// Computes the last content address of this Item.
        /// </summary>
        public StructID LastId => this.Length == 1 ? Id : new StructID(Id.Client, Id.Clock + Length - 1);

        /// <summary>
        /// Returns the next non-deleted item.
        /// </summary>
        public IItem Next
        {
            get
            {
                var n = Right;
                while (n != null && n.Deleted)
                {
                    n = (n as IItem)?.Right;
                }
                return n;
            }
        }

        /// <summary>
        /// Returns the previous non-deleted item.
        /// </summary>
        public IItem Prev
        {
            get
            {
                var n = Left;
                while (n != null && n.Deleted)
                {
                    n = (n as IItem)?.Left;
                }
                return n;
            }
        }

        public void MarkDeleted()
        {
            _info |= InfoEnum.Deleted;
        }

        /// <summary>
        /// Try to merge two items.
        /// </summary>
        /// https://github.com/yjs/yjs/blob/c2f0ca3faef731a93ba28f286ffafe0e1ea1a3aa/src/structs/Item.js#L557
        public  bool MergeWith(IItem right)
        {
            var rightItem = right as IItem;

            if (StructID.Equals(rightItem?.LeftOrigin, LastId) &&
                Right == right &&
                StructID.Equals(rightItem?.RightOrigin, RightOrigin) &&
                Id.Client == right.Id.Client &&
                Id.Clock + Length == right.Id.Clock &&
                Deleted == right.Deleted &&
                Redone == null &&
                rightItem?.Redone == null &&
                Content.GetType().IsAssignableFrom(rightItem?.Content?.GetType()) &&
                Content.MergeWith(rightItem.Content))
            {
                if (rightItem.Keep)
                {
                    Keep = true;
                }

                Right = rightItem.Right;
                if (Right is IItem)
                {
                    (Right as IItem).Left = this;
                }

                Length += rightItem.Length;
                return true;
            }

            return false;
        }

        /// <summary>
        /// Mark this item as Deleted.
        /// </summary>
        public  void Delete(ITransaction transaction)
        {
            if (!Deleted)
            {
                var parent = Parent as AbstractType;
                if (Countable && ParentSub == null)
                {
                    Debug.Assert(parent != null);
                    parent.Length -= Length;
                }

                MarkDeleted();
                transaction.DeleteSet.Add(Id.Client, Id.Clock, Length);
                transaction.AddChangedTypeToTransaction(parent, ParentSub);
                Content.Delete(transaction);
            }
        }

        public  void Integrate(ITransaction transaction, int offset)
        {
            if (offset > 0)
            {
                Id = new StructID(Id.Client, Id.Clock + offset);
                Left = transaction.Doc.Store.GetItemCleanEnd(transaction, new StructID(Id.Client, Id.Clock - 1));
                LeftOrigin = (Left as IItem)?.LastId;
                Content = (IContentEx)Content.Splice(offset);
                Length -= offset;
            }

            if (Parent != null)
            {
                if ((Left == null && (Right == null || (Right as IItem)?.Left != null)) || (Left != null && (Left as IItem)?.Right != Right))
                {
                    var left = Left;
                    IItem o;

                    // Set 'o' to the first conflicting item.
                    if (left != null)
                    {
                        o = (left as IItem)?.Right;
                    }
                    else if (ParentSub != null)
                    {
                        Debug.Assert(Parent is AbstractType);

                        IItem item = null;
                        (Parent as AbstractType)?.Map?.TryGetValue(ParentSub, out item);
                        o = (IItem)item;

                        while (o != null && (o as IItem)?.Left != null)
                        {
                            o = (o as IItem).Left;
                        }
                    }
                    else
                    {
                        Debug.Assert(Parent is AbstractType);
                        o = (IItem)(Parent as AbstractType)?.Start;
                    }

                    var conflictingItems = new HashSet<IItem>();
                    var itemsBeforeOrigin = new HashSet<IItem>();

                    while (o != null && o != Right)
                    {
                        itemsBeforeOrigin.Add(o);
                        conflictingItems.Add(o);

                        if (StructID.Equals(LeftOrigin, (o as IItem)?.LeftOrigin))
                        {
                            // Case 1
                            if (o.Id.Client < Id.Client)
                            {
                                left = o;
                                conflictingItems.Clear();
                            }
                            else if (StructID.Equals(RightOrigin, (o as IItem)?.RightOrigin))
                            {
                                // This and 'o' are conflicting and point to the same integration points.
                                // The id decides which item comes first.
                                // Since this is to the left of 'o', we can break here.
                                break;
                            }
                            // Else, 'o' might be integrated before an item that this conflicts with.
                            // If so, we will find it in the next iterations.
                        }
                        // Use 'Find' instead of 'GetItemCleanEnd', because we don't want / need to split items.
                        else if ((o as IItem)?.LeftOrigin != null && itemsBeforeOrigin.Contains(transaction.Doc.Store.Find((o as IItem).LeftOrigin.Value)))
                        {
                            // Case 2
                            // TODO: Store.Find is called twice here, call once?
                            if (!conflictingItems.Contains(transaction.Doc.Store.Find((o as IItem).LeftOrigin.Value)))
                            {
                                left = o;
                                conflictingItems.Clear();
                            }
                        }
                        else
                        {
                            break;
                        }

                        o = (o as IItem)?.Right;
                    }

                    Left = left;
                }

                // Reconnect left/right + update parent map/start if necessary.
                if (Left != null)
                {
                    if (Left is IItem leftItem)
                    {
                        var right = leftItem.Right;
                        Right = right;
                        leftItem.Right = this;
                    }
                    else
                    {
                        Right = null;
                    }
                }
                else
                {
                    IItem r;

                    if (ParentSub != null)
                    {
                        IItem item = null;
                        (Parent as AbstractType)?.Map?.TryGetValue(ParentSub, out item);
                        r = (IItem)item;

                        while (r != null && (r as IItem)?.Left != null)
                        {
                            r = (r as IItem).Left;
                        }
                    }
                    else
                    {
                        if (Parent is AbstractType abstractTypeParent)
                        {
                            r = (IItem)abstractTypeParent.Start;
                            abstractTypeParent.Start = this;
                        }
                        else
                        {
                            r = null;
                        }
                    }

                    Right = r;
                }

                if (Right != null)
                {
                    if (Right is IItem rightItem)
                    {
                        rightItem.Left = this;
                    }
                }
                else if (ParentSub != null)
                {
                    // Set as current parent value if right == null and this is parentSub.
                    (Parent as AbstractType).Map[ParentSub] = this;
                    // This is the current attribute value of parent. Delete right.
                    Left?.Delete(transaction);
                }

                // Adjust length of parent.
                if (ParentSub == null && Countable && !Deleted)
                {
                    Debug.Assert(Parent is AbstractType);
                    (Parent as AbstractType).Length += Length;
                }

                transaction.Doc.Store.AddStruct(this);
                Content.Integrate(transaction, this);

                // Add parent to transaction.changed.
                transaction.AddChangedTypeToTransaction(Parent as AbstractType, ParentSub);

                if (((Parent as AbstractType)?.Item != null && (Parent as AbstractType).Item.Deleted) || (ParentSub != null && Right != null))
                {
                    // Delete if parent is deleted or if this is not the current attribute value of parent.
                    Delete(transaction);
                }
            }
            else
            {
                // Parent is not defined. Integrate GC struct instead.
                new GC(Id, Length).Integrate(transaction, 0);
            }
        }

        /// <summary>
        /// Returns the creator ClientID of the missing OP or define missing items and return null.
        /// </summary>
        public  long? GetMissing(ITransaction transaction, IStructStore store)
        {
            if (LeftOrigin != null && LeftOrigin.Value.Client != Id.Client && LeftOrigin.Value.Clock >= store.GetState(LeftOrigin.Value.Client))
            {
                return LeftOrigin.Value.Client;
            }

            if (RightOrigin != null && RightOrigin.Value.Client != Id.Client && RightOrigin.Value.Clock >= store.GetState(RightOrigin.Value.Client))
            {
                return RightOrigin.Value.Client;
            }

            if ((Parent is StructID parentId) && Id.Client != parentId.Client && parentId.Clock >= store.GetState(parentId.Client))
            {
                return parentId.Client;
            }

            // We have all missing ids, now find the items.

            if (LeftOrigin != null)
            {
                Left = store.GetItemCleanEnd(transaction, LeftOrigin.Value);
                LeftOrigin = (Left as IItem)?.LastId;
            }

            if (RightOrigin != null)
            {
                Right = store.GetItemCleanStart(transaction, RightOrigin.Value);
                RightOrigin = Right.Id;
            }

            if (Left is GC || Right is GC)
            {
                Parent = null;
            }

            // Only set parent if this shouldn't be garbage collected.
            if (Parent == null)
            {
                if (Right is IItem rightItem)
                {
                    Parent = rightItem.Parent;
                    ParentSub = rightItem.ParentSub;
                }
                else if (Left is IItem leftItem)
                {
                    Parent = leftItem.Parent;
                    ParentSub = leftItem.ParentSub;
                }
            }
            else if (Parent is StructID pid)
            {
                var parentItem = store.Find(pid);
                if (parentItem is GC)
                {
                    Parent = null;
                }
                else
                {
                    Parent = ((parentItem as IItem)?.Content as ContentType)?.Type;
                }
            }

            return null;
        }

        public void Gc(IStructStore store, bool parentGCd)
        {
            if (!Deleted)
            {
                throw new InvalidOperationException();
            }

            Content.Gc(store);

            if (parentGCd)
            {
                store.ReplaceStruct(this, new GC(Id, Length));
            }
            else
            {
                Content = new ContentDeleted(Length);
            }
        }

        /// <summary>
        /// Make sure that neither item nor any of its parents is ever deleted.
        /// This property does not persist when storing it into a database or when
        /// sending it to other peers.
        /// </summary>
        public void KeepItemAndParents(bool value)
        {
            IItem item = this;

            while (item != null && item.Keep != value)
            {
                item.Keep = value;
                item = (item.Parent as AbstractType)?.Item;
            }
        }

        public bool IsVisible(ISnapshot snap)
        {
            return snap == null
                ? !Deleted
                : snap.StateVector.ContainsKey(Id.Client) && snap.StateVector[Id.Client] > Id.Clock && !snap.DeleteSet.IsDeleted(Id);
        }

        public  void Write(IUpdateEncoder encoder, int offset)
        {
            var origin = offset > 0 ? new StructID(Id.Client, Id.Clock + offset - 1) : LeftOrigin;
            var rightOrigin = RightOrigin;
            var parentSub = ParentSub;
            var info = (Content.Ref & Bits.Bits5) |
                (origin == null ? 0 : Bit.Bit8) |
                (rightOrigin == null ? 0 : Bit.Bit7) |
                (parentSub == null ? 0 : Bit.Bit6);
            encoder.WriteInfo((byte)info);

            if (origin != null)
            {
                encoder.WriteLeftId(origin.Value);
            }

            if (rightOrigin != null)
            {
                encoder.WriteRightId(rightOrigin.Value);
            }

            if (origin == null && rightOrigin == null)
            {
                var parent = Parent;
                var parentItem = (parent as AbstractType)?.Item;
                if (parentItem == null)
                {
                    // parent type on y._map.
                    // find the correct key
                    var yKey = (parent as AbstractType).FindRootTypeKey();
                    encoder.WriteParentInfo(true);
                    encoder.WriteString(yKey);
                }
                else
                {
                    encoder.WriteParentInfo(false);
                    encoder.WriteLeftId(parentItem.Id);
                }

                if (parentSub != null)
                {
                    encoder.WriteString(parentSub);
                }
            }

            Content.Write(encoder, offset);
        }

        /// <summary>
        /// Split 'leftItem' into two items.
        /// </summary>
        public IItem SplitItem(ITransaction transaction, int diff)
        {
            var client = Id.Client;
            var clock = Id.Clock;

            var rightItem = new Item(
                new StructID(client, clock + diff),
                this,
                new StructID(client, clock + diff - 1),
                Right,
                RightOrigin,
                Parent,
                ParentSub,
                Content.Splice(diff));

            if (Deleted)
            {
                rightItem.MarkDeleted();
            }

            if (Keep)
            {
                rightItem.Keep = true;
            }

            if (Redone != null)
            {
                rightItem.Redone = new StructID(Redone.Value.Client, Redone.Value.Clock + diff);
            }

            // Update left (do not set leftItem.RightOrigin as it will lead to problems when syncing).
            Right = rightItem;

            // Update right.
            if (rightItem.Right is IItem rightIt)
            {
                rightIt.Left = rightItem;
            }

            // Right is more specific.
            transaction.MergeStructs.Add(rightItem);

            // Update parent._map.
            if (rightItem.ParentSub != null && rightItem.Right == null)
            {
                (rightItem.Parent as AbstractType).Map[rightItem.ParentSub] = rightItem;
            }

            Length = diff;
            return rightItem;
        }
    }
}
